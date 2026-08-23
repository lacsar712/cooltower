package app

import (
	"context"

	"github.com/lacsar712/cooltower/internal/model"
)

func (a *App) BeginCycleScope(ctx context.Context, tower model.TowerID) (context.Context, context.CancelFunc) {
	if tower == "" {
		tower = model.TowerID(a.cfg.TowerID)
	}
	a.cycleMu.Lock()
	if cancel, ok := a.cycleCancels[tower]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.cycleCancels[tower] = cancel
	a.cycleMu.Unlock()
	release := func() {
		a.cycleMu.Lock()
		delete(a.cycleCancels, tower)
		a.cycleMu.Unlock()
		cancel()
	}
	return child, release
}

type CycleOptions struct {
	Tower model.TowerID
}

func (a *App) RunCycle(ctx context.Context, opt CycleOptions) error {
	tower := opt.Tower
	if tower == "" {
		tower = model.TowerID(a.cfg.TowerID)
	}
	cycleCtx, release := a.BeginCycleScope(ctx, tower)
	defer release()
	if err := a.towerFSM.Apply(cycleCtx, "prime"); err != nil {
		return err
	}
	if err := a.unit.Prime(cycleCtx); err != nil {
		return err
	}
	return a.towerFSM.Apply(cycleCtx, "spray_ok")
}
