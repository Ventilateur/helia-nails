package treatwell

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	twmodels "github.com/Ventilateur/helia-nails/treatwell/models"
)

func (tw *Treatwell) List(_ context.Context, employee models.Employee, from time.Time, to time.Time) ([]models.Appointment, error) {
	twCalendar, err := tw.getCalendar(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}

	var appointments []models.Appointment
	for _, twAppointment := range twCalendar.Appointments {
		if twAppointment.EmployeeId != employee.Treatwell.Id {
			continue
		}

		appointments = append(appointments, twAppointment.CoreModel(tw.config))
	}

	slices.SortFunc(appointments, func(a1, a2 models.Appointment) int {
		return a1.StartTime.Compare(a2.StartTime)
	})

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
