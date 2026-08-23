package tower

import (
	"context"
	"sync"
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

type Basin struct {
	mu          sync.RWMutex
	id          model.BasinID
	levelPct    float64
	temperature float64
	filledAt    time.Time
}

func NewBasin(id model.BasinID, initialLevel float64) *Basin {
	return &Basin{id: id, levelPct: initialLevel, temperature: 22.0}
}

func (b *Basin) ID() model.BasinID { return b.id }

func (b *Basin) Level() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.levelPct
}

func (b *Basin) Temperature() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.temperature
}

func (b *Basin) SetTemperature(celsius float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.temperature = celsius
}

func (b *Basin) Fill(ctx context.Context, at time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	select {
	case <-ctx.Done():
		return model.Wrap("basin", "canceled", context.Cause(ctx))
	default:
	}
	if b.levelPct >= 95 {
		return model.Wrap("basin", "full", model.ErrConflict)
	}
	b.levelPct = 90
	b.filledAt = at
	return nil
}

func (b *Basin) Draw(gpm float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.levelPct <= 5 {
		return model.Wrap("basin", "low", model.ErrConflict)
	}
	b.levelPct -= gpm * 0.01
	if b.levelPct < 5 {
		b.levelPct = 5
	}
	return nil
}

func (b *Basin) Cool(delta float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.temperature -= delta
	if b.temperature < 10 {
		b.temperature = 10
	}
}

func (b *Basin) FilledAt() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.filledAt
}
