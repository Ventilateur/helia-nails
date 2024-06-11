package classpass

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/utils"
	"google.golang.org/api/calendar/v3"
)

const (
	BookEventName  = "__SYNC__"
	BlockEventName = "__BLOCKED_BY_TREATWELL__"
)

func (c *Classpass) Book(_ context.Context, appointment models.Appointment) error {
	event := &calendar.Event{
		Summary:     BookEventName,
		Location:    string(appointment.Source),
		Description: appointment.CustomNotes(),
		Start: &calendar.EventDateTime{
			DateTime: appointment.StartTime.In(time.UTC).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: appointment.EndTime.In(time.UTC).Format(time.RFC3339),
		},
	}

	_, err := c.svc.Events.Insert(appointment.Employee.Classpass.GoogleCalendarId, event).Do()
	if err != nil {
		return fmt.Errorf("failed to book event %s: %w", appointment, err)
	}

	return nil
}

func (c *Classpass) Block(_ context.Context, employee models.Employee, from, to time.Time) error {
	events, err := c.svc.Events.List(employee.Classpass.GoogleCalendarId).
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
		Location: string(models.SourceTreatwell),
		Start: &calendar.EventDateTime{
			DateTime: from.In(time.UTC).Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: to.In(time.UTC).Format(time.RFC3339),
		},
	}

	_, err = c.svc.Events.Insert(employee.Classpass.GoogleCalendarId, event).Do()
	if err != nil {
		return fmt.Errorf("failed to block time slot from %s to %s: %w", from, to, err)
	}

	return nil
}
