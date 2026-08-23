package alarms

import (
	"context"
	"sync"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/model"
)

type Emitter struct {
	mu      sync.Mutex
	reg     *Registry
	clk     clock.Clock
	bufSize int
	active  []model.AlarmEvent
	history []model.AlarmEvent
}

func NewEmitter(reg *Registry, clk clock.Clock, bufSize int) *Emitter {
	if bufSize <= 0 {
		bufSize = 32
	}
	return &Emitter{reg: reg, clk: clk, bufSize: bufSize}
}

func (e *Emitter) Raise(ctx context.Context, code model.AlarmCode, tower model.TowerID, severity int) error {
	select {
	case <-ctx.Done():
		return model.Wrap("alarms", "canceled", context.Cause(ctx))
	default:
	}
	ev := model.AlarmEvent{
		Code: code, Message: e.reg.Message(code), Tower: tower,
		RaisedAt: e.clk.Now(), Severity: severity,
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.active = append(e.active, ev)
	e.history = append(e.history, ev)
	if len(e.history) > e.bufSize {
		e.history = e.history[len(e.history)-e.bufSize:]
	}
	return nil
}

func (e *Emitter) Active() []model.AlarmEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]model.AlarmEvent, len(e.active))
	copy(out, e.active)
	return out
}

func (e *Emitter) History() []model.AlarmEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]model.AlarmEvent, len(e.history))
	copy(out, e.history)
	return out
}

func (e *Emitter) Clear(code model.AlarmCode) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var kept []model.AlarmEvent
	for _, ev := range e.active {
		if ev.Code != code {
			kept = append(kept, ev)
		}
	}
	e.active = kept
}

func (e *Emitter) ClearAll() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.active = nil
}
