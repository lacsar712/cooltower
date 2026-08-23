package fan

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/fsm"
	"github.com/lacsar712/cooltower/internal/model"
)

type Bank struct {
	mu    sync.RWMutex
	id    model.FanID
	vfd   *VFD
	fsm   *fsm.FanFSM
	clk   clock.Clock
	speed float64
}

func NewBank(id model.FanID, clk clock.Clock, minRun, coast time.Duration) *Bank {
	b := &Bank{id: id, clk: clk, vfd: NewVFD(id, minRun, coast)}
	b.fsm = fsm.NewFanFSM(id, b.onTransition)
	return b
}

func (b *Bank) onTransition(ctx context.Context, fan model.FanID, from, to model.FanState) error {
	switch to {
	case model.FanRun:
		return b.vfd.Enable(ctx)
	case model.FanOff, model.FanCoast:
		return b.vfd.Disable(ctx)
	case model.FanTrip:
		return b.vfd.Trip(ctx)
	}
	return nil
}

func (b *Bank) ID() model.FanID { return b.id }

func (b *Bank) State() model.FanState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.fsm.State()
}

func (b *Bank) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := b.fsm.Apply(ctx, "start"); err != nil {
		return err
	}
	return b.fsm.Apply(ctx, "staged")
}

func (b *Bank) Stop(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.fsm.State() != model.FanRun {
		return model.Wrap("fan_bank", "not_running", model.ErrConflict)
	}
	if err := b.fsm.Apply(ctx, "stop"); err != nil {
		return err
	}
	deadline := b.clk.Now().Add(b.vfd.CoastDuration())
	for b.clk.Now().Before(deadline) {
		if pc, ok := b.clk.(*clock.ProcessClock); ok {
			pc.Step()
		} else {
			break
		}
	}
	return b.fsm.Apply(ctx, "coast_done")
}

func (b *Bank) SetSpeed(pct float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.fsm.State() != model.FanRun {
		return model.Wrap("fan_bank", "not_running", model.ErrConflict)
	}
	if pct < 0 || pct > 100 {
		return model.Wrap("fan_bank", "speed_range", model.ErrInvalidID)
	}
	b.speed = pct
	return b.vfd.SetTarget(pct)
}

func (b *Bank) Speed() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.speed
}

func (b *Bank) Assignment() model.FanAssignment {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return model.FanAssignment{
		Fan: b.id, SpeedPct: b.speed, Enabled: b.fsm.State() == model.FanRun, State: b.fsm.State(),
	}
}

func (b *Bank) StatusLine() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return fmt.Sprintf("fan=%s state=%s speed=%.0f%%", b.id, b.fsm.State(), b.speed)
}

type Coordinator struct {
	mu    sync.RWMutex
	banks map[model.FanID]*Bank
}

func NewCoordinator() *Coordinator {
	return &Coordinator{banks: make(map[model.FanID]*Bank)}
}

func (c *Coordinator) Add(b *Bank) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.banks[b.ID()] = b
}

func (c *Coordinator) StartAll(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for id, bank := range c.banks {
		if err := bank.Start(ctx); err != nil {
			return model.Wrap("fan_coord", string(id), err)
		}
	}
	return nil
}

func (c *Coordinator) AverageSpeed() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.banks) == 0 {
		return 0
	}
	var sum float64
	for _, b := range c.banks {
		sum += b.Speed()
	}
	return sum / float64(len(c.banks))
}

func (c *Coordinator) Assignments() []model.FanAssignment {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]model.FanAssignment, 0, len(c.banks))
	for _, b := range c.banks {
		out = append(out, b.Assignment())
	}
	return out
}

func (c *Coordinator) Bank(id model.FanID) (*Bank, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	b, ok := c.banks[id]
	return b, ok
}

func (c *Coordinator) Banks() []*Bank {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*Bank, 0, len(c.banks))
	for _, b := range c.banks {
		out = append(out, b)
	}
	return out
}
