package treatwell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/mapping"
	twmodels "github.com/Ventilateur/helia-nails/treatwell/models"
	"github.com/Ventilateur/helia-nails/utils"
)

const (
	baseURL = "https://connect.treatwell.fr"
	apiURL  = baseURL + "/api"
)

type Treatwell struct {
	httpClient *http.Client

	venueID              string
	employees            *twmodels.Employees
	employeeWorkingHours *twmodels.EmployeesWorkingHours
}

func New(httpClient *http.Client, venueID string) (*Treatwell, error) {
	return &Treatwell{
		httpClient: httpClient,
		venueID:    venueID,
	}, nil
}

func (tw *Treatwell) Preload(from, to time.Time) error {
	var err error

	tw.employees, err = tw.GetEmployeesInfo()
	if err != nil {
		return fmt.Errorf("failed to get employees info: %w", err)
	}

	tw.employeeWorkingHours, err = tw.GetAllEmployeesWorkingHours(from, to)
	if err != nil {
		return fmt.Errorf("failed to get employees working hours: %w", err)
	}

	return nil
}

func (tw *Treatwell) bootstrapCookie() error {
	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return utils.RequestCreationErr(baseURL, err)
	}

	req.Header.Add("authority", "connect.treatwell.fr")
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("upgrade-insecure-requests", "1")

	resp, err := tw.httpClient.Do(req)
	if err != nil {
		return utils.DoRequestErr(baseURL, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return utils.UnexpectedErrorCode(resp.StatusCode)
	}

	return nil
}

func (tw *Treatwell) Login(user, password string) error {
	err := tw.bootstrapCookie()
	if err != nil {
		return err
	}

	payload := strings.NewReader(fmt.Sprintf(
		`{"user":"%s","password":"%s","persistentLogin":false}`,
		user, password,
	))

	req, err := http.NewRequest(http.MethodPost, apiURL+"/authentication.json", payload)
	if err != nil {
		return utils.RequestCreationErr(apiURL+"/authentication.json", err)
	}

	writeCommonHeaders(req, map[string]string{
		"referer": "https://connect.treatwell.fr/login",
	})

	resp, err := tw.httpClient.Do(req)
	if err != nil {
		return utils.DoRequestErr(apiURL+"/authentication.json", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return utils.UnexpectedErrorCode(resp.StatusCode)
	}

	return nil
}

func (tw *Treatwell) ListAppointments(from, to time.Time) (map[string]models.Appointment, error) {
	twCalendar, err := tw.GetCalendar(from, to)
	if err != nil {
		return nil, err
	}

	appointments := map[string]models.Appointment{}
	for _, appointment := range twCalendar.Appointments {
		start, end, err := utils.ParseFromToTimes(
			fmt.Sprintf("%sT%s:00", appointment.AppointmentDate, appointment.StartTime),
			fmt.Sprintf("%sT%s:00", appointment.AppointmentDate, appointment.EndTime),
		)
		if err != nil {
			return nil, err
		}

		source, id := utils.ParseCustomID(appointment.Notes)
		if id == "" {
			id = strconv.Itoa(appointment.Id)
		}

		appointments[id] = models.Appointment{
			Id:        id,
			Source:    source,
			Employee:  appointment.EmployeeName,
			StartTime: start,
			EndTime:   end,
			Offer:     appointment.OfferName,
			Notes:     appointment.Notes,
		}
	}

	return appointments, nil
}

type BookAppointmentsRequest struct {
	Appointments    []twmodels.Appointment `json:"appointments"`
	VenueCustomerID *int                   `json:"venueCustomerId"`
	AnonymousNote   *string                `json:"anonymousNote"`
}

func (tw *Treatwell) BookAnonymously(appointment models.Appointment) error {
	offer, ok := mapping.TreatWellOffers[appointment.Offer]
	if !ok {
		return fmt.Errorf("no Treatwell offer found for [%s]", appointment.Offer)
	}

	twAppointment := &twmodels.Appointment{
		AppointmentDate: appointment.StartTime.Format(time.DateOnly),
		StartTime:       fmt.Sprintf("%02d:%02d", appointment.StartTime.Hour(), appointment.StartTime.Minute()),
		EndTime:         fmt.Sprintf("%02d:%02d", appointment.EndTime.Hour(), appointment.EndTime.Minute()),
		Platform:        "DESKTOP",
		Notes:           fmt.Sprintf("${%s:%s}", string(appointment.Source), appointment.Id),
		ServiceId:       offer.OfferID,
		Skus: []twmodels.Sku{
			{
				SkuId: offer.SkuID,
			},
		},
	}

	calendar, err := tw.GetCalendar(appointment.StartTime, appointment.EndTime)
	if err != nil {
		return fmt.Errorf("failed to get calendar: %w", err)
	}

	err = tw.bookAnonymously(twAppointment, appointment.ClientName, calendar)
	if err != nil {
		return fmt.Errorf("failed to book TW: %w", err)
	}

	return nil
}

func findBookableEmployee(
	appointment *twmodels.Appointment,
	employees *twmodels.Employees,
	employeeWorkingHours *twmodels.EmployeesWorkingHours,
	calendar *twmodels.Calendar,
) (employeeID int, slotFound bool) {
	for _, employee := range employees.Employees {
		canOffer := slices.Contains(employee.EmployeeOffers, appointment.ServiceId)
		if !canOffer {
			continue
		}

		// Employee can offer service
		overlapped := false
		for _, employeesWorkingHour := range employeeWorkingHours.EmployeesWorkingHours {
			if employeesWorkingHour.EmployeeID == employee.Id {
				for _, workingHour := range employeesWorkingHour.WorkingHours {
					if workingHour.Date == appointment.AppointmentDate && len(workingHour.TimeSlots) > 0 &&
						workingHour.TimeSlots[0].TimeFrom <= appointment.StartTime &&
						workingHour.TimeSlots[0].TimeTo >= appointment.EndTime {

						// Employee works at the requested hour
						for _, bookedAppointment := range calendar.Appointments {
							if bookedAppointment.AppointmentDate == appointment.AppointmentDate && bookedAppointment.EmployeeId == employee.Id {
								overlapped = isOverlapping(bookedAppointment, *appointment)
								if overlapped {
									// Overlapped booking
									break
								}
							}
						}

						if !overlapped {
							return employee.Id, true
						}
					}
				}
				break
			}
		}
	}

	return 0, false
}

func isOverlapping(a, b twmodels.Appointment) bool {
	return a.StartAt().Compare(b.EndAt()) < 0 && a.EndAt().Compare(b.StartAt()) > 0
}

func (tw *Treatwell) bookAnonymously(appointment *twmodels.Appointment, clientName string, calendar *twmodels.Calendar) error {
	employeeID, slotFound := findBookableEmployee(appointment, tw.employees, tw.employeeWorkingHours, calendar)
	if !slotFound {
		return fmt.Errorf("no slot found for service %d at [%s-%s] on %s",
			appointment.ServiceId,
			appointment.StartTime,
			appointment.EndTime,
			appointment.AppointmentDate,
		)
	}

	appointment.EmployeeId = employeeID

	reqBody := &BookAppointmentsRequest{
		Appointments:    []twmodels.Appointment{*appointment},
		VenueCustomerID: nil,
		AnonymousNote:   &clientName,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("%s: %w", utils.ErrUnmarshalJSON, err)
	}

	return doRequestWithoutResponse(
		tw,
		http.MethodPost,
		apiURL+"/venue/"+tw.venueID+"/appointments",
		bytes.NewBuffer(payload),
		nil,
	)
}

func (tw *Treatwell) GetCalendar(fromDate, toDate time.Time) (*twmodels.Calendar, error) {
	return doRequestWithResponse[twmodels.Calendar](
		tw,
		http.MethodGet,
		apiURL+"/venue/"+tw.venueID+"/calendar.json",
		nil,
		map[string]string{
			"include":                  "appointments",
			"appointment-status-codes": "CR,CN,NS,CP",
			"utm_source":               "calendar-regular",
			"date-from":                fromDate.Format(time.DateOnly),
			"date-to":                  toDate.Format(time.DateOnly),
		},
	)
}

func (tw *Treatwell) GetAllEmployeesWorkingHours(from, to time.Time) (*twmodels.EmployeesWorkingHours, error) {
	var employeeIDs []string
	for _, employee := range tw.employees.Employees {
		employeeIDs = append(employeeIDs, strconv.Itoa(employee.Id))
	}

	return doRequestWithResponse[twmodels.EmployeesWorkingHours](
		tw,
		http.MethodGet,
		apiURL+"/venue/"+tw.venueID+"/employees/"+strings.Join(employeeIDs, ",")+"/working-hours.json",
		nil,
		map[string]string{
			"from": from.Format(time.DateOnly),
			"to":   to.Format(time.DateOnly),
		},
	)
}

func (tw *Treatwell) GetEmployeesInfo() (*twmodels.Employees, error) {
	return doRequestWithResponse[twmodels.Employees](
		tw,
		http.MethodGet,
		apiURL+"/venue/"+tw.venueID+"/employees.json",
		nil,
		map[string]string{
			"include": "employee-offers",
		},
	)
}

func writeCommonHeaders(req *http.Request, extra map[string]string) {
	req.Header.Add("authority", "connect.treatwell.fr")
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("origin", "https://connect.treatwell.fr")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("x-requested-with", "XMLHttpRequest")

	for k, v := range extra {
		req.Header.Add(k, v)
	}
}

func doRequestWithResponse[T any](tw *Treatwell, method, u string, body io.Reader, params map[string]string) (*T, error) {
	resp, err := tw.doRequest(method, u, body, params)
	if err != nil {
		return nil, utils.DoRequestErr(u, err)
	}
	defer resp.Body.Close()

	ok := func() bool {
		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated, http.StatusNoContent:
			return true
		default:
			return false
		}
	}()

	if !ok {
		return nil, utils.UnexpectedErrorCode(resp.StatusCode)
	}

	var ret T
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", utils.ErrUnmarshalJSON, err)
	}

	return &ret, nil
}

func doRequestWithoutResponse(tw *Treatwell, method, u string, body io.Reader, params map[string]string) error {
	resp, err := tw.doRequest(method, u, body, params)
	if err != nil {
		return utils.DoRequestErr(u, err)
	}
	defer resp.Body.Close()

	ok := func() bool {
		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated, http.StatusNoContent:
			return true
		default:
			return false
		}
	}()

	if !ok {
		return utils.UnexpectedErrorCode(resp.StatusCode)
	}

	return nil
}

func (tw *Treatwell) doRequest(method, u string, body io.Reader, params map[string]string) (*http.Response, error) {
	urlObj, err := url.Parse(u)
	if err != nil {
		return nil, utils.URLParseErr(u, err)
	}

	q := urlObj.Query()
	for k, v := range params {
		q.Set(k, v)
	}

	urlObj.RawQuery = q.Encode()

	req, err := http.NewRequest(method, urlObj.String(), body)
	if err != nil {
		return nil, utils.RequestCreationErr(urlObj.String(), err)
	}

	writeCommonHeaders(req, map[string]string{})

	return tw.httpClient.Do(req)
}
