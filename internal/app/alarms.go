package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/cooltower/internal/model"
)

func (a *App) ReportDriftFault(ctx context.Context, ppm float64) error {
	if ppm > a.cfg.DriftMaxPPM {
		_ = a.alarms.Raise(ctx, "DRIFT_HIGH", a.TowerID(), 3)
		return fmt.Errorf("tower fault: %w", model.ErrDriftExceeded)
	}
	return nil
}

func (a *App) HandleFillBlock(ctx context.Context, delta float64) error {
	if err := a.fillGuard.Permit(delta); err != nil {
		_ = a.alarms.Raise(ctx, "FILL_BLOCK", a.TowerID(), 2)
		return fmt.Errorf("tower: %w", err)
	}
	return nil
}
