package googlecalendar

import (
	"fmt"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"google.golang.org/api/calendar/v3"
)

func (c *GoogleCalendar) Book(calendarID string, appointment models.Appointment) error {
	_, err := c.svc.Events.Insert(
		calendarID,
		&calendar.Event{
			Id:          appointment.Id,
			Summary:     appointment.Offer,
			Location:    string(appointment.Source),
			Description: appointment.Notes,
			Start: &calendar.EventDateTime{
				DateTime: appointment.StartTime.Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: appointment.EndTime.Format(time.RFC3339),
			},
		},
	).Do()

	if err != nil {
		return fmt.Errorf("failed to book event %s: %w", appointment.Id, err)
	}

	return nil
}
