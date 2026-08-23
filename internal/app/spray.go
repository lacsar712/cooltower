package app

import (
	"context"
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

func (a *App) ExecutePlan(ctx context.Context, entries []model.SprayScheduleEntry) error {
	return a.spraySched.InstallSprayPlanCtx(ctx, entries)
}

func (a *App) SchedulerItemCount() int {
	return a.spraySched.ItemCount()
}

func (a *App) RunAntifreezeScheduler(ctx context.Context, anchor time.Time) error {
	if err := a.antifreeze.Require(anchor); err != nil {
		return err
	}
	entries := []model.SprayScheduleEntry{{
		Header: model.SprayHeaderID(a.cfg.TowerID + "-spray-01"),
		Start:  anchor, End: anchor.Add(time.Hour),
		Setpoint: model.FlowSetpoint{GallonsPerMinute: a.cfg.DefaultSprayGPM, TolerancePct: a.cfg.FlowTolerancePct},
	}}
	return a.ExecutePlan(ctx, entries)
}
