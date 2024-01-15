package treatwell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/mapping"
	twmodels "github.com/Ventilateur/helia-nails/treatwell/models"
	"github.com/Ventilateur/helia-nails/utils"
)

type BookAppointmentsRequest struct {
	Appointments    []twmodels.Appointment `json:"appointments"`
	VenueCustomerID *int                   `json:"venueCustomerId"`
	AnonymousNote   *string                `json:"anonymousNote"`
}

func (tw *Treatwell) BookAnonymously(appointment models.Appointment) error {
	offer := func() *mapping.TreatwellOffer {
		for _, twOffer := range mapping.TreatwellOffers {
			for _, name := range twOffer.PossibleNames {
				if strings.Contains(appointment.Offer, name) {
					return &twOffer
				}
			}
		}
		return nil
	}()
	if offer == nil {
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

	calendar, err := tw.getCalendar(appointment.StartTime, appointment.EndTime)
	if err != nil {
		return fmt.Errorf("failed to get calendar: %w", err)
	}

	err = tw.bookAnonymously(
		twAppointment,
		fmt.Sprintf("[%s] %s", appointment.Source, appointment.ClientName),
		calendar,
	)
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
