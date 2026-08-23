package interlock

import (
	"context"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/model"
)

type DriftGuard struct {
	budget model.DriftBudget
	clk    clock.Clock
	hold   *holdState
}

type holdState struct {
	active  bool
	started time.Time
	until   time.Time
}

func NewDriftGuard(clk clock.Clock, budget model.DriftBudget) *DriftGuard {
	return &DriftGuard{clk: clk, budget: budget}
}

func (g *DriftGuard) PermitSpray(ctx context.Context, driftPPM, fanSpeedPct, sprayGPM float64) error {
	if g.budget.Exceeded(driftPPM, fanSpeedPct, sprayGPM) {
		g.armHold()
		return model.Wrap("drift_guard", "budget_exceeded", model.ErrInterlock)
	}
	if g.hold != nil && g.hold.active && g.clk.Now().Before(g.hold.until) {
		return model.Wrap("drift_guard", "hold_active", model.ErrDriftHold)
	}
	return nil
}

func (g *DriftGuard) PermitFanSpeed(ctx context.Context, driftPPM, fanSpeedPct, sprayGPM float64) error {
	if g.budget.Exceeded(driftPPM, fanSpeedPct, sprayGPM) {
		return model.Wrap("drift_guard", "fan_speed", model.ErrInterlock)
	}
	return nil
}

func (g *DriftGuard) armHold() {
	now := g.clk.Now()
	dur := g.budget.HoldDuration
	if dur <= 0 {
		dur = 2 * time.Minute
	}
	g.hold = &holdState{active: true, started: now, until: now.Add(dur)}
}

func (g *DriftGuard) HoldActive() bool {
	if g.hold == nil || !g.hold.active {
		return false
	}
	if g.clk.Now().After(g.hold.until) {
		g.hold.active = false
		return false
	}
	return true
}

func (g *DriftGuard) ClearHold() {
	if g.hold != nil {
		g.hold.active = false
	}
}

func (g *DriftGuard) Budget() model.DriftBudget { return g.budget }

type FanSprayPair struct {
	Fan   model.FanID
	Spray model.SprayHeaderID
}

type Guard struct {
	allowed map[model.FanID]model.SprayHeaderID
}

func NewGuard(pairs map[model.FanID]model.SprayHeaderID) *Guard {
	cp := make(map[model.FanID]model.SprayHeaderID, len(pairs))
	for k, v := range pairs {
		cp[k] = v
	}
	return &Guard{allowed: cp}
}

func (g *Guard) Permit(fan model.FanID, spray model.SprayHeaderID) error {
	want, ok := g.allowed[fan]
	if !ok {
		return model.Wrap("interlock", "unknown_fan", model.ErrNotFound)
	}
	if want != spray {
		return model.Wrap("interlock", "spray_mismatch", model.ErrInterlock)
	}
	return nil
}

func (g *Guard) SpraysFor(fan model.FanID) (model.SprayHeaderID, bool) {
	spray, ok := g.allowed[fan]
	return spray, ok
}

func (g *Guard) Pairs() []FanSprayPair {
	out := make([]FanSprayPair, 0, len(g.allowed))
	for fan, spray := range g.allowed {
		out = append(out, FanSprayPair{Fan: fan, Spray: spray})
	}
	return out
}
