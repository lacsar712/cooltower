package tower

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/model"
)

type Unit struct {
	mu      sync.RWMutex
	id      model.TowerID
	basin   *Basin
	clk     clock.Clock
	driftPPM float64
}

func NewUnit(id model.TowerID, basin *Basin, clk clock.Clock) *Unit {
	return &Unit{id: id, basin: basin, clk: clk}
}

func (u *Unit) ID() model.TowerID {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.id
}

func (u *Unit) Basin() *Basin {
	return u.basin
}

func (u *Unit) SetDriftReading(ppm float64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.driftPPM = ppm
}

func (u *Unit) DriftPPM() float64 {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.driftPPM
}

func (u *Unit) BasinTemperature() float64 {
	if u.basin == nil {
		return 0
	}
	return u.basin.Temperature()
}

func (u *Unit) Prime(ctx context.Context) error {
	if u.basin == nil {
		return model.Wrap("tower", "no_basin", model.ErrNotFound)
	}
	return u.basin.Fill(ctx, u.clk.Now())
}

func (u *Unit) Snapshot(state model.TowerState, fans []model.FanAssignment, sprays []model.SprayAssignment) model.TowerSnapshot {
	u.mu.RLock()
	defer u.mu.RUnlock()
	fanCopy := make([]model.FanAssignment, len(fans))
	copy(fanCopy, fans)
	sprayCopy := make([]model.SprayAssignment, len(sprays))
	copy(sprayCopy, sprays)
	temp := 0.0
	if u.basin != nil {
		temp = u.basin.Temperature()
	}
	return model.TowerSnapshot{
		ID: u.id, State: state, Fans: fanCopy, Sprays: sprayCopy,
		DriftPPM: u.driftPPM, BasinTemp: temp, UpdatedAt: u.clk.Now(),
	}
}

func (u *Unit) StatusLine(state model.TowerState, fanCount, sprayCount int) string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return fmt.Sprintf("tower=%s state=%s drift=%.1fppm basin=%.1fC fans=%d sprays=%d",
		u.id, state, u.driftPPM, u.BasinTemperature(), fanCount, sprayCount)
}

type Registry struct {
	mu     sync.RWMutex
	units  map[model.TowerID]*Unit
}

func NewRegistry() *Registry {
	return &Registry{units: make(map[model.TowerID]*Unit)}
}

func (r *Registry) Register(u *Unit) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.units[u.ID()] = u
}

func (r *Registry) Get(id model.TowerID) (*Unit, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.units[id]
	return u, ok
}

func (r *Registry) List() []*Unit {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Unit, 0, len(r.units))
	for _, u := range r.units {
		out = append(out, u)
	}
	return out
}
