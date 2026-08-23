package app

import (
	"context"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/fan"
	"github.com/lacsar712/cooltower/internal/model"
)

type FanRamp struct {
	clk   clock.Clock
	steps int
	delay time.Duration
}

func NewFanRamp(clk clock.Clock, steps int, delay time.Duration) *FanRamp {
	if steps <= 0 {
		steps = 40
	}
	if delay <= 0 {
		delay = time.Millisecond
	}
	return &FanRamp{clk: clk, steps: steps, delay: delay}
}

func (r *FanRamp) Ramp(ctx context.Context, coord *fan.Coordinator) error {
	for _, bank := range coord.Banks() {
		// Honor an operator abort between staged banks: the cascade must stop
		// accepting new stages once the cancel signal has propagated.
		if err := ctx.Err(); err != nil {
			return model.Wrap("fan_ramp", "canceled", err)
		}
		if err := bank.Start(ctx); err != nil {
			return err
		}
		for step := 0; step < r.steps; step++ {
			if err := ctx.Err(); err != nil {
				return model.Wrap("fan_ramp", "canceled", err)
			}
			target := float64(step+1) * 100 / float64(r.steps)
			if err := bank.SetSpeed(target); err != nil {
				return err
			}
			if pc, ok := r.clk.(*clock.ProcessClock); ok {
				pc.Step()
			}
			// Wait between speed steps, but bail out the instant the operator
			// revokes the ramp instead of running the remaining steps to full
			// speed. time.Sleep would ignore the cancel entirely.
			select {
			case <-ctx.Done():
				return model.Wrap("fan_ramp", "canceled", context.Cause(ctx))
			case <-time.After(r.delay):
			}
		}
	}
	return nil
}

func (a *App) FanRamp(ctx context.Context) error {
	return a.fanRamp.Ramp(ctx, a.fanCoord)
}
