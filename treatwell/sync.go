package treatwell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	twmodels "github.com/Ventilateur/helia-nails/treatwell/models"
	"github.com/Ventilateur/helia-nails/utils"
)

const (
	baseURL = "https://connect.treatwell.fr"
	apiURL  = baseURL + "/api"
)

type Treatwell struct {
	httpClient *http.Client

	venueID   string
	employees []twmodels.Employee
}

func New(httpClient *http.Client, venueID string) (*Treatwell, error) {
	return &Treatwell{
		httpClient: httpClient,
		venueID:    venueID,
	}, nil
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

func (tw *Treatwell) GetCalendar(from, to time.Time) (*twmodels.Calendar, error) {
	return doRequestWithResponse[twmodels.Calendar](
		tw,
		http.MethodGet,
		apiURL+"/venue/"+tw.venueID+"/calendar.json",
		nil,
		map[string]string{
			"include":                  "appointments",
			"appointment-status-codes": "CR,CN,NS,CP",
			"utm_source":               "calendar-regular",
			"date-from":                from.Format(time.DateOnly),
			"date-to":                  to.Format(time.DateOnly),
		},
	)
}

type BookAppointmentsRequest struct {
	Appointments    []twmodels.Appointment `json:"appointments"`
	VenueCustomerID *int                   `json:"venueCustomerId"`
	AnonymousNote   *string                `json:"anonymousNote"`
}

func (tw *Treatwell) BookAnonymously(
	appointment twmodels.Appointment,
	clientName string,
	employees twmodels.Employees,
	employeeWorkingHours twmodels.EmployeesWorkingHours,
	calendar twmodels.Calendar,
) error {
	slotFound := false
	employeeId := 0
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
					if workingHour.Date == appointment.AppointmentDate &&
						len(workingHour.TimeSlots) > 0 &&
						workingHour.TimeSlots[0].TimeFrom <= appointment.StartTime &&
						workingHour.TimeSlots[0].TimeTo <= appointment.EndTime {

						// Employee works at the requested hour
						for _, bookedAppointment := range calendar.Appointments {
							if bookedAppointment.AppointmentDate == appointment.AppointmentDate {
								if bookedAppointment.StartTime < appointment.EndTime || appointment.StartTime < bookedAppointment.EndTime {
									// Overlapped booking
									overlapped = true
									break
								}
							}
						}

						break
					}
				}
				break
			}
		}

		if !overlapped {
			slotFound = true
			employeeId = employee.Id
		}
	}

	if !slotFound {
		return fmt.Errorf("no slot found for service %q at [%s-%s] on %s",
			appointment.ServiceId,
			appointment.StartTime,
			appointment.EndTime,
			appointment.AppointmentDate,
		)
	}

	appointment.EmployeeId = employeeId

	reqBody := &BookAppointmentsRequest{
		Appointments:    []twmodels.Appointment{appointment},
		VenueCustomerID: nil,
		AnonymousNote:   &clientName,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return utils.ErrUnmarshalJSON
	}

	return doRequestWithoutResponse(
		tw,
		http.MethodPost,
		apiURL+"/venue/"+tw.venueID+"/appointments",
		bytes.NewBuffer(payload),
		nil,
	)
}

func (tw *Treatwell) GetEmployeeWorkingHours(employeeIDs []string, from, to time.Time) (*twmodels.EmployeesWorkingHours, error) {
	for _, employee := range tw.employees {
		employeeIDs = append(employeeIDs, fmt.Sprintf("%d", employee.Id))
	}

	return doRequestWithResponse[twmodels.EmployeesWorkingHours](
		tw,
		http.MethodGet,
		apiURL+"/venue/"+tw.venueID+"/employees/"+strings.Join(employeeIDs, ",")+"/working-hours.json",
		nil,
		map[string]string{
			"date-from": from.Format(time.DateOnly),
			"date-to":   to.Format(time.DateOnly),
		},
	)
}

func (tw *Treatwell) GetEmployeeInfo() (*twmodels.Employees, error) {
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

	if resp.StatusCode != http.StatusOK {
		return nil, utils.UnexpectedErrorCode(resp.StatusCode)
	}

	var ret T
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, utils.ErrUnmarshalJSON
	}

	return &ret, nil
}

func doRequestWithoutResponse(tw *Treatwell, method, u string, body io.Reader, params map[string]string) error {
	resp, err := tw.doRequest(method, u, body, params)
	if err != nil {
		return utils.DoRequestErr(u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
