package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Ventilateur/helia-nails/aws"
	"github.com/Ventilateur/helia-nails/classpass"
	"github.com/Ventilateur/helia-nails/config"
	"github.com/Ventilateur/helia-nails/planity"
	"github.com/Ventilateur/helia-nails/treatwell"
	"github.com/Ventilateur/helia-nails/utils"
	"github.com/aws/aws-lambda-go/lambda"
	"gopkg.in/yaml.v3"
)

type Event struct {
	Name                   string `json:"name"`
	PlanityAccessTokenPath string `json:"planityAccessTokenPath"`
	PlatformConfigPath     string `json:"platformConfigPath"`
	ConfigFileBucket       string `json:"configFileBucket"`
	ConfigFilePath         string `json:"configFilePath"`
}

func HandleRequest(ctx context.Context, event *Event) (*string, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}

	switch event.Name {
	case "sync":
		params, err := aws.GetParam(event.PlatformConfigPath, event.PlanityAccessTokenPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get parameter: %w", err)
		}

		configFile, err := aws.GetConfig(event.ConfigFileBucket, event.ConfigFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get config file: %w", err)
		}

		// Merge platforms' global config file in parameter store (more sensitive)
		// with the config file in S3 (less sensitive)
		fullConfig := []byte(fmt.Sprintf("%s\n\n%s", params[event.PlatformConfigPath], configFile))
		cfg := &config.Config{}

		if err := yaml.Unmarshal(fullConfig, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}

		cfg.Planity.AccessToken = params[event.PlanityAccessTokenPath]

		tw, err := treatwell.New(&http.Client{Timeout: 1 * time.Minute}, cfg)
		if err != nil {
			return nil, err
		}

		cp, err := classpass.New(ctx, cfg)
		if err != nil {
			return nil, err
		}

		pl, err := planity.New(ctx, &http.Client{Timeout: 15 * time.Second}, cfg)
		if err != nil {
			return nil, err
		}

		if pl.AccessToken() != params[event.PlanityAccessTokenPath] {
			if err := aws.SetParam(event.PlanityAccessTokenPath, pl.AccessToken()); err != nil {
				return nil, err
			}
		}

		sync := Sync{
			tw:  tw,
			cp:  cp,
			pl:  pl,
			cfg: cfg,
		}

		from := utils.BoD(time.Now())
		to := utils.EoD(from.Add(14 * 24 * time.Hour))
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
