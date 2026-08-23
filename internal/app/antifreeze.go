package app

import (
	"context"
	"time"
)

// ConfirmAntifreeze reports whether the antifreeze保温窗 has elapsed.
// When the window has not yet closed it MUST surface the antifreeze hold
// (ErrAntifreezeHold), not a generic "schedule empty" — otherwise the HMI
// cannot distinguish "保温窗未到点，继续等" from "排程被清空，需补排程".
func (a *App) ConfirmAntifreeze(ctx context.Context, anchor time.Time) error {
	_ = ctx
	return a.antifreeze.Require(anchor)
}

func (a *App) AntifreezeReady(anchor time.Time) bool {
	return a.antifreeze.Ready(anchor)
}
