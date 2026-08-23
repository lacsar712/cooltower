package spray

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/fsm"
	"github.com/lacsar712/cooltower/internal/model"
)

type Header struct {
	mu       sync.RWMutex
	id       model.SprayHeaderID
	fsm      *fsm.SprayFSM
	clk      clock.Clock
	setpoint model.FlowSetpoint
	flow     float64
	nozzles  []*Nozzle
}

func NewHeader(id model.SprayHeaderID, clk clock.Clock, nozzleCount int, setpoint model.FlowSetpoint) (*Header, error) {
	if nozzleCount <= 0 {
		return nil, model.Wrap("spray_header", "nozzle_count", model.ErrInvalidID)
	}
	h := &Header{id: id, clk: clk, setpoint: setpoint}
	h.fsm = fsm.NewSprayFSM(id, h.onTransition)
	for i := 0; i < nozzleCount; i++ {
		nzID, err := model.ParseNozzleID(id, i)
		if err != nil {
			return nil, err
		}
		h.nozzles = append(h.nozzles, NewNozzle(nzID))
	}
	return h, nil
}

func (h *Header) onTransition(ctx context.Context, header model.SprayHeaderID, from, to model.SprayState) error {
	switch to {
	case model.SprayActive:
		return h.openNozzles(ctx)
	case model.SprayClosed:
		return h.closeNozzles()
	case model.SprayThrottled:
		return h.syncFlow()
	case model.SprayFault:
		return h.closeNozzles()
	}
	return nil
}

func (h *Header) openNozzles(ctx context.Context) error {
	for _, nz := range h.nozzles {
		if err := nz.Open(ctx); err != nil {
			return err
		}
	}
	return h.syncFlow()
}

func (h *Header) closeNozzles() error {
	for _, nz := range h.nozzles {
		nz.Close()
	}
	h.flow = 0
	return nil
}

func (h *Header) syncFlow() error {
	var total float64
	for _, nz := range h.nozzles {
		total += nz.Flow()
	}
	h.flow = total
	return nil
}

func (h *Header) ID() model.SprayHeaderID { return h.id }

func (h *Header) State() model.SprayState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.fsm.State()
}

func (h *Header) Open(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.fsm.Apply(ctx, "open"); err != nil {
		return err
	}
	return h.fsm.Apply(ctx, "flow_ok")
}

func (h *Header) Throttle(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.fsm.Apply(ctx, "throttle")
}

func (h *Header) Close(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.fsm.Apply(ctx, "close")
}

func (h *Header) Flow() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.flow
}

func (h *Header) Setpoint() model.FlowSetpoint {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.setpoint
}

func (h *Header) BindSetpoint(sp model.FlowSetpoint) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.setpoint = sp
}

func (h *Header) Assignment() model.SprayAssignment {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return model.SprayAssignment{
		Header: h.id, Setpoint: h.setpoint, Enabled: h.fsm.State() == model.SprayActive,
		LastFlow: h.flow, State: h.fsm.State(),
	}
}

func (h *Header) StatusLine() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return fmt.Sprintf("spray=%s state=%s flow=%.1fgpm", h.id, h.fsm.State(), h.flow)
}

type Plant struct {
	mu      sync.RWMutex
	headers map[model.SprayHeaderID]*Header
}

func NewPlant() *Plant {
	return &Plant{headers: make(map[model.SprayHeaderID]*Header)}
}

func (p *Plant) Add(h *Header) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.headers[h.ID()] = h
}

func (p *Plant) Get(id model.SprayHeaderID) (*Header, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	h, ok := p.headers[id]
	return h, ok
}

func (p *Plant) TotalFlow() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var total float64
	for _, h := range p.headers {
		total += h.Flow()
	}
	return total
}

func (p *Plant) OpenAll(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for id, h := range p.headers {
		if err := h.Open(ctx); err != nil {
			return model.Wrap("spray_plant", string(id), err)
		}
	}
	return nil
}

func (p *Plant) Assignments() []model.SprayAssignment {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]model.SprayAssignment, 0, len(p.headers))
	for _, h := range p.headers {
		out = append(out, h.Assignment())
	}
	return out
}

func (p *Plant) ValidateFlows(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for id, h := range p.headers {
		if h.State() != model.SprayActive {
			continue
		}
		if !h.Setpoint().Within(h.Flow()) {
			return model.Wrap("spray_plant", string(id), model.ErrFlowSetpoint)
		}
	}
	return nil
}
