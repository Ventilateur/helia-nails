package classpass

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/utils"
	"google.golang.org/api/calendar/v3"
)

var (
	classPassGoogleCalendarSummary = regexp.MustCompile(`([^:]+):([^:]+)\(ClassPass Booking\)`)
)

func (c *Classpass) List(ctx context.Context, employee models.Employee, from time.Time, to time.Time) ([]models.Appointment, error) {
	events, err := c.svc.Events.List(employee.Classpass.GoogleCalendarId).
		SingleEvents(true).
		ShowDeleted(false).
		MaxResults(1000).
		TimeMin(from.Format(time.RFC3339)).
		TimeMax(to.Format(time.RFC3339)).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events for %s: %w", employee.Classpass.GoogleCalendarId, err)
	}

	var appointments []models.Appointment
	for _, event := range events.Items {
		if event.Summary == BlockEventName {
			continue
		}

		appointment, err := c.parseGoogleEvent(event)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Google event: %w", err)
		}
		appointment.Employee = employee
		appointments = append(appointments, appointment)
	}

	return appointments, nil
}

func (c *Classpass) parseGoogleEvent(event *calendar.Event) (models.Appointment, error) {
	start, end, err := utils.ParseTimes(event.Start.DateTime, event.End.DateTime)
	if err != nil {
		return models.Appointment{}, fmt.Errorf("failed to parse from/to times: %w", err)
	}

	// Created by ClassPass
	matches := classPassGoogleCalendarSummary.FindStringSubmatch(event.Summary)
	if len(matches) == 3 {
		return models.Appointment{
			Source: models.SourceClassPass,
			Ids: models.AppointmentIds{
				Classpass: event.Id,
			},
			Service:    c.config.GetService(models.SourceClassPass, strings.TrimSpace(matches[2]), ""),
			StartTime:  start,
			EndTime:    end,
			ClientName: matches[1],
		}, nil
	}

	source, id := utils.ParseCustomID(event.Description)
	if source == "" {
		return models.Appointment{}, fmt.Errorf("unknown source")
	}

	// Created by sync
	return models.Appointment{
		Source: source, // Should NEVER be Classpass
		Ids: models.AppointmentIds{
			Classpass: event.Id,
			Treatwell: func() string {
				if source == models.SourceTreatwell {
					return id
				}
				return ""
			}(),
			Planity: func() string {
				if source == models.SourcePlanity {
					return id
				}
				return ""
			}(),
		},
		Service:   models.Service{}, // No need because it's a blocking event on Classpass POV
		StartTime: start,
		EndTime:   end,
		Notes:     event.Description,
	}, nil
}
