package clock

import (
	"fmt"
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

type AntifreezeWindow struct {
	clk      Clock
	duration time.Duration
}

func NewAntifreezeWindow(clk Clock, duration time.Duration) *AntifreezeWindow {
	if duration <= 0 {
		duration = 5 * time.Minute
	}
	return &AntifreezeWindow{clk: clk, duration: duration}
}

func (w *AntifreezeWindow) Ready(startedAt time.Time) bool {
	return WindowClosed(w.clk, startedAt, w.duration)
}

func (w *AntifreezeWindow) Require(startedAt time.Time) error {
	if w.Ready(startedAt) {
		return nil
	}
	return fmt.Errorf("antifreeze window: %w", model.ErrAntifreezeHold)
}
