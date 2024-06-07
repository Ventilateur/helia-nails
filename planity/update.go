package planity

import (
	"context"
	"fmt"

	coremodels "github.com/Ventilateur/helia-nails/core/models"
)

func (p *Planity) Update(ctx context.Context, appointment coremodels.Appointment) error {
	if err := p.Delete(ctx, appointment); err != nil {
		return fmt.Errorf("failed to update appointment %s: %w", appointment, err)
	}

	if err := p.Book(ctx, appointment); err != nil {
		return fmt.Errorf("failed to update appointment %s: %w", appointment, err)
	}

	return nil
}
