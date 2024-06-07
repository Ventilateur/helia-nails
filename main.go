package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Ventilateur/helia-nails/aws"
	"github.com/Ventilateur/helia-nails/config"
	"github.com/Ventilateur/helia-nails/utils"
	"github.com/aws/aws-lambda-go/lambda"
	"gopkg.in/yaml.v3"
)

const (
	configPath = "config"
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
		params, err := aws.GetParam(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get parameter: %w", err)
		}

		cfg := &config.Config{}
		if err := yaml.Unmarshal([]byte(params[configPath]), cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}

		from := utils.BoD(time.Now())
		to := utils.EoD(from.Add(7 * 24 * time.Hour))
		if err := syncAll(ctx, cfg, from, to); err != nil {
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
