package classpass

import (
	"context"
	"fmt"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"google.golang.org/api/calendar/v3"
)

func (c *Classpass) Update(_ context.Context, appointment models.Appointment) error {
	_, err := c.svc.Events.Patch(
		appointment.Employee.Classpass.GoogleCalendarId,
		appointment.Ids.Classpass,
		&calendar.Event{
			Start: &calendar.EventDateTime{
				DateTime: appointment.StartTime.Format(time.RFC3339),
			},
			End: &calendar.EventDateTime{
				DateTime: appointment.EndTime.Format(time.RFC3339),
			},
		},
	).Do()
	if err != nil {
		return fmt.Errorf("failed to update event %s: %w", appointment, err)
	}

	return nil
}