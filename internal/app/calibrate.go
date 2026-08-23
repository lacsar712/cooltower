package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/cooltower/internal/model"
)

// CalibrateProbe allows acceptance tests to inject calibration sensor faults.
var CalibrateProbe func(ctx context.Context) error

func (a *App) CalibrateSpray(ctx context.Context, segment model.SprayHeaderID, holder string) error {
	lease, err := a.segmentLeases.Acquire(string(segment), holder, a.clk.Now())
	if err != nil {
		return err
	}
	defer lease.Release()
	if CalibrateProbe != nil {
		if err := CalibrateProbe(ctx); err != nil {
			return fmt.Errorf("calibrate: %w", err)
		}
	}
	return nil
}

func (a *App) SegmentHeld(segment model.SprayHeaderID) bool {
	return a.segmentLeases.IsHeld(string(segment))
}
