package treatwell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	twmodels "github.com/Ventilateur/helia-nails/treatwell/models"
	"github.com/Ventilateur/helia-nails/utils"
)

type BookAppointmentsRequest struct {
	Appointments    []twmodels.Appointment `json:"appointments"`
	VenueCustomerID *int                   `json:"venueCustomerId"`
	AnonymousNote   *string                `json:"anonymousNote"`
}

func (tw *Treatwell) Book(_ context.Context, appointment models.Appointment) error {
	twAppointment := &twmodels.Appointment{
		AppointmentDate: appointment.StartTime.Format(time.DateOnly),
		StartTime:       fmt.Sprintf("%02d:%02d", appointment.StartTime.Hour(), appointment.StartTime.Minute()),
		EndTime:         fmt.Sprintf("%02d:%02d", appointment.EndTime.Hour(), appointment.EndTime.Minute()),
		Platform:        "DESKTOP",
		EmployeeId:      appointment.Employee.Treatwell.Id,
		Notes:           appointment.CustomNotes(),
		ServiceId:       appointment.Service.Treatwell.OfferId,
		Skus: []twmodels.Sku{
			{
				SkuId: appointment.Service.Treatwell.SkuId,
			},
		},
	}

	//calendar, err := tw.getCalendar(appointment.StartTime, appointment.EndTime)
	//if err != nil {
	//	return fmt.Errorf("failed to get calendar: %w", err)
	//}

	employeeInfo := tw.employeeInfo[appointment.Employee.Treatwell.Id]

	// Double check if employee can perform a service
	canOffer := slices.Contains(employeeInfo.Info.EmployeeOffers, twAppointment.ServiceId)
	if !canOffer {
		return fmt.Errorf("employee %s cannot perform '%s'", appointment.Employee.Name, appointment.Service.Name)
	}

	// Check for valid time slot and overlapping
	//for _, workingHour := range employeeInfo.WorkingHours {
	//	if workingHour.Date == twAppointment.AppointmentDate && // this day
	//		len(workingHour.TimeSlots) > 0 && // employee works
	//		workingHour.TimeSlots[0].TimeFrom <= twAppointment.StartTime && // working hour starts before the appointment
	//		workingHour.TimeSlots[0].TimeTo >= twAppointment.EndTime { // working hour ends after the appointment
	//
	//		// If employee works at the requested hour, check if there are already booked appointments there
	//		for _, bookedAppointment := range calendar.Appointments {
	//			if bookedAppointment.AppointmentDate == twAppointment.AppointmentDate && bookedAppointment.EmployeeId == employeeInfo.Info.Id {
	//				if isOverlapping(bookedAppointment, *twAppointment) {
	//					return fmt.Errorf(
	//						"employee %s cannot perform at %s %s because of overlapping",
	//						appointment.Employee.Name,
	//						twAppointment.AppointmentDate,
	//						twAppointment.StartTime,
	//					)
	//				}
	//			}
	//		}
	//	}
	//}

	return tw.book(twAppointment, fmt.Sprintf("[%s] %s", appointment.Source, appointment.ClientName))
}

func isOverlapping(a, b twmodels.Appointment) bool {
	return a.StartAt().Compare(b.EndAt()) < 0 && a.EndAt().Compare(b.StartAt()) > 0
}

func (tw *Treatwell) book(appointment *twmodels.Appointment, clientName string) error {
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

func (tw *Treatwell) Block(_ context.Context, _ models.Employee, _, _ time.Time) error {
	panic("not implemented")
}
