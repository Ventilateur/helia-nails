package treatwell

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	twmodels "github.com/Ventilateur/helia-nails/treatwell/models"
	"github.com/Ventilateur/helia-nails/utils"
)

func (tw *Treatwell) ListAppointments(employee string, from, to time.Time) (map[string]models.Appointment, error) {
	twCalendar, err := tw.getCalendar(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}

	appointments := map[string]models.Appointment{}
	for _, appointment := range twCalendar.Appointments {
		if appointment.EmployeeName != employee {
			continue
		}

		start, end, err := utils.ParseTimes(
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
			Id:               id,
			Source:           source,
			Employee:         appointment.EmployeeName,
			ClientName:       appointment.ConsumerName,
			StartTime:        start,
			EndTime:          end,
			Offer:            appointment.OfferName,
			Notes:            appointment.Notes,
			OriginalPlatform: models.PlatformTreatwell,
			OriginalID:       strconv.Itoa(appointment.Id),
		}
	}

	return appointments, nil
}

func (tw *Treatwell) getCalendar(fromDate, toDate time.Time) (*twmodels.Calendar, error) {
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
