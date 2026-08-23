package app

import (
	"context"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/fan"
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
		if err := bank.Start(ctx); err != nil {
			return err
		}
		for step := 0; step < r.steps; step++ {
			select {
			case <-ctx.Done():
				return context.Cause(ctx)
			default:
			}
			target := float64(step+1) * 100 / float64(r.steps)
			if err := bank.SetSpeed(target); err != nil {
				return err
			}
			if pc, ok := r.clk.(*clock.ProcessClock); ok {
				pc.Step()
			}
			time.Sleep(r.delay)
		}
	}
	return nil
}

func (a *App) FanRamp(ctx context.Context) error {
	return a.fanRamp.Ramp(ctx, a.fanCoord)
}
