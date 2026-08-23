package app

import (
	"context"
	"fmt"
	"time"
)

func (a *App) ConfirmAntifreeze(ctx context.Context, anchor time.Time) error {
	if err := a.antifreeze.Require(anchor); err != nil {
		return fmt.Errorf("schedule: %w", err)
	}
	_ = ctx
	return nil
}

func (a *App) AntifreezeReady(anchor time.Time) bool {
	return a.antifreeze.Ready(anchor)
}
