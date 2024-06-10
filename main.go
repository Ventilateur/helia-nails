package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Ventilateur/helia-nails/aws"
	"github.com/Ventilateur/helia-nails/config"
	"github.com/Ventilateur/helia-nails/planity"
	"github.com/Ventilateur/helia-nails/treatwell"
	"github.com/Ventilateur/helia-nails/utils"
	"github.com/aws/aws-lambda-go/lambda"
	"gopkg.in/yaml.v3"
)

const (
	configFile         = "config"
	planityAccessToken = "/planity/accessToken"
)

type Event struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, event *Event) (*string, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	switch event.Name {
	case "sync":
		params, err := aws.GetParam(configFile, planityAccessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get parameter: %w", err)
		}

		cfg := &config.Config{}
		if err := yaml.Unmarshal([]byte(params[configFile]), cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}

		cfg.Planity.AccessToken = params[planityAccessToken]

		tw, err := treatwell.New(&http.Client{Timeout: 1 * time.Minute}, cfg)
		if err != nil {
			return nil, err
		}

		//cp, err := classpass.New(ctx, cfg)
		//if err != nil {
		//	return err
		//}

		pl, err := planity.New(ctx, &http.Client{Timeout: 15 * time.Second}, cfg)
		if err != nil {
			return nil, err
		}

		sync := Sync{
			tw:  tw,
			cp:  nil,
			pl:  pl,
			cfg: cfg,
		}

		from := utils.BoD(time.Now())
		to := utils.EoD(from.Add(7 * 24 * time.Hour))
		if err := sync.syncAll(ctx, from, to); err != nil {
			return nil, fmt.Errorf("failed to sync: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid event name %s", event.Name)
	}

	return nil, nil
}

func main() {
	lambda.Start(HandleRequest)
}
