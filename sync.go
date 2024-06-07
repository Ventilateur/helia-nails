package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Ventilateur/helia-nails/classpass"
	"github.com/Ventilateur/helia-nails/config"
	"github.com/Ventilateur/helia-nails/core"
	"github.com/Ventilateur/helia-nails/planity"
	"github.com/Ventilateur/helia-nails/treatwell"
)

func syncAll(ctx context.Context, cfg *config.Config, from time.Time, to time.Time) error {
	tw, err := treatwell.New(&http.Client{Timeout: 1 * time.Minute}, cfg)
	if err != nil {
		return err
	}

	cp, err := classpass.New(ctx, cfg)
	if err != nil {
		return err
	}

	pl, err := planity.New(ctx, &http.Client{Timeout: 15 * time.Second}, cfg)
	if err != nil {
		return err
	}

	if err := tw.Preload(from, to); err != nil {
		return fmt.Errorf("failed to preload TW data: %w", err)
	}

	for _, employee := range cfg.Employees {
		for _, platform := range []core.Platform{cp, pl} {
			if err := core.SyncWorkingHours(cfg, tw, employee, from, to, platform); err != nil {
				return fmt.Errorf("failed to sync working hours to %s: %w", platform.Name(), err)
			}
			if err := core.Sync(tw, platform, employee, from, to); err != nil {
				return fmt.Errorf("failed to sync %s <-> %s: %w", tw.Name(), platform.Name(), err)
			}
		}
	}

	return nil
}
