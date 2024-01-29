package googlecalendar

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/utils"
	"google.golang.org/api/calendar/v3"
)

const (
	defaultIANATz = "Europe/Paris"

	BlockEventName = "__BLOCKED_BY_TREATWELL__"
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

func (c *GoogleCalendar) Block(calendarID string, from, to time.Time) error {
	events, err := c.svc.Events.List(calendarID).
		TimeMin(from.Format(time.RFC3339)).
		TimeMax(to.Format(time.RFC3339)).
		MaxResults(1000).
		Do()
	if err != nil {
		return fmt.Errorf("failed to list events: %w", err)
	}

	for _, item := range events.Items {
		if item.Summary == BlockEventName {
			start, end, err := utils.ParseTimes(item.Start.DateTime, item.End.DateTime)
			if err != nil {
				return fmt.Errorf("failed to parse event date time: %w", err)
			}
			if start.Truncate(time.Hour).Equal(from.Truncate(time.Hour)) &&
				end.Truncate(time.Hour).Equal(to.Truncate(time.Hour)) {
				slog.Info(fmt.Sprintf("Already blocked from %s to %s", start, end))
				return nil
			}
		}
	}

	event := &calendar.Event{
		Summary:  BlockEventName,
		Location: string(models.PlatformTreatwell),
		Start: &calendar.EventDateTime{
			DateTime: from.Format(time.RFC3339),
			//TimeZone: defaultIANATz,
		},
		End: &calendar.EventDateTime{
			DateTime: to.Format(time.RFC3339),
			//TimeZone: defaultIANATz,
		},
	}

	_, err = c.svc.Events.Insert(calendarID, event).Do()
	if err != nil {
		return fmt.Errorf("failed to block time slot from %s to %s: %w", from, to, err)
	}

	return nil
}
