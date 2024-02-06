package googlecalendar

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/mapping"
	"github.com/Ventilateur/helia-nails/utils"
	"google.golang.org/api/calendar/v3"
)

var (
	classPassGoogleCalendarSummary = regexp.MustCompile(`([^:]+):([^:]+)\(ClassPass Booking\)`)
)

func (c *GoogleCalendar) List(calendarID string, from time.Time, to time.Time) (map[string]models.Appointment, error) {
	events, err := c.svc.Events.List(calendarID).
		SingleEvents(true).
		ShowDeleted(false).
		MaxResults(1000).
		TimeMin(from.Format(time.RFC3339)).
		TimeMax(to.Format(time.RFC3339)).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events for %s: %w", calendarID, err)
	}

	employee := mapping.CalendarIDToEmployeeMap[calendarID]

	appointments := map[string]models.Appointment{}
	for _, item := range events.Items {
		if item.Summary == BlockEventName {
			continue
		}

		appointment, err := parseGoogleEvent(item)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Google event: %w", err)
		}
		appointment.Employee = employee

		appointments[appointment.Id] = *appointment
	}

	return appointments, nil
}

func parseGoogleEvent(event *calendar.Event) (*models.Appointment, error) {
	start, end, err := utils.ParseTimes(event.Start.DateTime, event.End.DateTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse from/to times: %w", err)
	}

	_, id := utils.ParseCustomID(event.Description)
	if id == "" {
		id = event.Id
	}

	// Created by ClassPass
	matches := classPassGoogleCalendarSummary.FindStringSubmatch(event.Summary)
	if len(matches) == 3 {
		return &models.Appointment{
			Id:               id,
			Source:           models.SourceClassPass,
			Employee:         "Unknown",
			StartTime:        start,
			EndTime:          end,
			Offer:            strings.TrimSpace(matches[2]),
			ClientName:       matches[1],
			OriginalPlatform: models.PlatformGoogle,
			OriginalID:       event.Id,
		}, nil
	}

	// Created by sync
	return &models.Appointment{
		Id:               id,
		Source:           models.Source(event.Location),
		Employee:         "Unknown",
		StartTime:        start,
		EndTime:          end,
		Offer:            event.Summary,
		Notes:            event.Description,
		OriginalPlatform: models.PlatformGoogle,
		OriginalID:       event.Id,
	}, nil
}
