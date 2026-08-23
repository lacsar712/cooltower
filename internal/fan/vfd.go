package fan

import (
	"context"
	"sync"
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

type VFD struct {
	mu       sync.Mutex
	id       model.FanID
	enabled  bool
	tripped  bool
	target   float64
	actual   float64
	minRun   time.Duration
	coast    time.Duration
	started  time.Time
}

func NewVFD(id model.FanID, minRun, coast time.Duration) *VFD {
	return &VFD{id: id, minRun: minRun, coast: coast}
}

func (v *VFD) Enable(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	select {
	case <-ctx.Done():
		return model.Wrap("vfd", "canceled", context.Cause(ctx))
	default:
	}
	if v.tripped {
		return model.Wrap("vfd", "tripped", model.ErrFanFault)
	}
	v.enabled = true
	v.started = time.Now()
	v.actual = 30
	return nil
}

func (v *VFD) Disable(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	select {
	case <-ctx.Done():
		return model.Wrap("vfd", "canceled", context.Cause(ctx))
	default:
	}
	v.enabled = false
	v.actual = 0
	v.target = 0
	return nil
}

func (v *VFD) Trip(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.tripped = true
	v.enabled = false
	v.actual = 0
	return nil
}

func (v *VFD) SetTarget(pct float64) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.enabled {
		return model.Wrap("vfd", "disabled", model.ErrConflict)
	}
	v.target = pct
	v.actual = pct
	return nil
}

func (v *VFD) Actual() float64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.actual
}

func (v *VFD) Target() float64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.target
}

func (v *VFD) CoastDuration() time.Duration {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.coast
}

func (v *VFD) Running() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.enabled && !v.tripped
}
