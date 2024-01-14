package google_calendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Ventilateur/helia-nails/core/models"
	"github.com/Ventilateur/helia-nails/utils"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type GoogleCalendar struct {
	svc *calendar.Service
}

func New(ctx context.Context, email string, key []byte) (*GoogleCalendar, error) {
	conf := &jwt.Config{
		Email:      email,
		PrivateKey: key,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.events",
		},
		TokenURL: google.JWTTokenURL,
		Subject:  email,
	}

	client := conf.Client(ctx)

	svc, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to initiate client: %w", err)
	}

	return &GoogleCalendar{
		svc: svc,
	}, nil
}

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

func (c *GoogleCalendar) List(calendarID string, from time.Time, to time.Time) (map[string]models.Appointment, error) {
	events, err := c.svc.Events.List(calendarID).
		SingleEvents(true).
		MaxResults(1000).
		TimeMin(from.Format(time.RFC3339)).
		TimeMax(to.Format(time.RFC3339)).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events for %s: %w", calendarID, err)
	}

	appointments := map[string]models.Appointment{}
	for _, item := range events.Items {
		start, end, err := utils.ParseFromToTimes(item.Start.DateTime, item.End.DateTime)
		if err != nil {
			return nil, err
		}

		if strings.Contains(item.Summary, "ClassPass Booking") {

		}

		appointments[item.Id] = models.Appointment{
			Id:        item.Id,
			Source:    models.Source(item.Location),
			Employee:  "Unknown",
			StartTime: start,
			EndTime:   end,
			Offer:     item.Summary,
			Notes:     item.Description,
		}
	}

	return appointments, nil
}

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

func (c *GoogleCalendar) DeleteAppointment(calendarID string, id string) error {
	return c.svc.Events.Delete(calendarID, id).Do()
}
