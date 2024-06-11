package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Ventilateur/helia-nails/classpass"
	"github.com/Ventilateur/helia-nails/config"
	"github.com/Ventilateur/helia-nails/core"
	"github.com/Ventilateur/helia-nails/planity"
	"github.com/Ventilateur/helia-nails/treatwell"
)

type Sync struct {
	tw  *treatwell.Treatwell
	cp  *classpass.Classpass
	pl  *planity.Planity
	cfg *config.Config
}

func (s *Sync) syncAll(ctx context.Context, from time.Time, to time.Time) error {
	if err := s.tw.Preload(from, to); err != nil {
		return fmt.Errorf("failed to preload TW data: %w", err)
	}

	for _, employee := range s.cfg.Employees {
		for _, platform := range []core.Platform{s.pl, s.cp} {
			if err := core.SyncWorkingHours(s.cfg, s.tw, employee, from, to, platform); err != nil {
				return fmt.Errorf("failed to sync working hours to %s: %w", platform.Name(), err)
			}
			if err := core.Sync(ctx, s.tw, platform, employee, from, to); err != nil {
				return fmt.Errorf("failed to sync %s <-> %s: %w", s.tw.Name(), platform.Name(), err)
			}
		}
	}

	return nil
}
