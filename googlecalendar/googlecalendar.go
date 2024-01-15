package googlecalendar

import (
	"context"
	"fmt"

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
