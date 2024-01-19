package googlecalendar

import (
	"fmt"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"google.golang.org/api/calendar/v3"
)

const (
	defaultIANATz = "Europe/Paris"
)

func (c *GoogleCalendar) Book(calendarID string, appointment models.Appointment) error {
	event := &calendar.Event{
		Summary:     appointment.Offer,
		Location:    string(appointment.Source),
		Description: fmt.Sprintf("${%s:%s}\n%s", string(appointment.Source), appointment.Id, appointment.Notes),
		Start: &calendar.EventDateTime{
			DateTime: appointment.StartTime.Format(time.RFC3339),
			//TimeZone: defaultIANATz,
		},
		End: &calendar.EventDateTime{
			DateTime: appointment.EndTime.Format(time.RFC3339),
			//TimeZone: defaultIANATz,
		},
	}

	_, err := c.svc.Events.Insert(calendarID, event).Do()
	if err != nil {
		return fmt.Errorf("failed to book event %s: %w", appointment, err)
	}

	return nil
}
