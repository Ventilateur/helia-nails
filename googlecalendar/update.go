package googlecalendar

import (
	"fmt"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"google.golang.org/api/calendar/v3"
)

func (c *GoogleCalendar) Update(calendarID string, eventID string, appointment models.Appointment) error {
	_, err := c.svc.Events.Patch(calendarID, eventID, &calendar.Event{
		Start: &calendar.EventDateTime{
			DateTime: appointment.StartTime.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: appointment.EndTime.Format(time.RFC3339),
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to update event %s: %w", eventID, err)
	}

	return nil
}
