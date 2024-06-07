package classpass

import (
	"context"
	"fmt"

	"github.com/Ventilateur/helia-nails/config"
	"github.com/Ventilateur/helia-nails/core/models"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Classpass struct {
	svc    *calendar.Service
	config *config.Config
}

func New(ctx context.Context, config *config.Config) (*Classpass, error) {
	jwtConf := &jwt.Config{
		Email:      config.Classpass.GoogleEmail,
		PrivateKey: []byte(config.Classpass.GoogleKey),
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.events",
		},
		TokenURL: google.JWTTokenURL,
		Subject:  config.Classpass.GoogleEmail,
	}

	client := jwtConf.Client(ctx)

	svc, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to initiate client: %w", err)
	}

	return &Classpass{
		svc:    svc,
		config: config,
	}, nil
}

func (c *Classpass) Name() models.Source {
	return models.SourceClassPass
}
