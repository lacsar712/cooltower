package spray

import (
	"context"
	"sync"

	"github.com/lacsar712/cooltower/internal/model"
)

type Nozzle struct {
	mu     sync.Mutex
	id     model.NozzleID
	open   bool
	flow   float64
	rating float64
}

func NewNozzle(id model.NozzleID) *Nozzle {
	return &Nozzle{id: id, rating: 3.75}
}

func (n *Nozzle) ID() model.NozzleID { return n.id }

func (n *Nozzle) Open(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	select {
	case <-ctx.Done():
		return model.Wrap("nozzle", "canceled", context.Cause(ctx))
	default:
	}
	n.open = true
	n.flow = n.rating
	return nil
}

func (n *Nozzle) Close() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.open = false
	n.flow = 0
}

func (n *Nozzle) Throttle(pct float64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if !n.open {
		return
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	n.flow = n.rating * (pct / 100)
}

func (n *Nozzle) Flow() float64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.flow
}

func (n *Nozzle) IsOpen() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.open
}
