package interlock

import (
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

type SegmentLease struct {
	segment  string
	holder   string
	registry *SegmentLeaseRegistry
	released bool
}

func (l *SegmentLease) Release() {
	if l.registry == nil || l.released {
		return
	}
	l.registry.release(l.segment, l.holder)
	l.released = true
}

type SegmentLeaseRegistry struct {
	mu      sync.Mutex
	leases  map[string]string
	expires map[string]time.Time
	ttl     time.Duration
}

func NewSegmentLeaseRegistry(ttl time.Duration) *SegmentLeaseRegistry {
	return &SegmentLeaseRegistry{
		leases: make(map[string]string), expires: make(map[string]time.Time), ttl: ttl,
	}
}

func (r *SegmentLeaseRegistry) Acquire(segment, holder string, now time.Time) (*SegmentLease, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.leases[segment]; ok {
		if exp, ok := r.expires[segment]; ok && now.Before(exp) {
			return nil, fmt.Errorf("%s held by %s: %w", segment, h, model.ErrInterlock)
		}
	}
	r.leases[segment] = holder
	r.expires[segment] = now.Add(r.ttl)
	return &SegmentLease{segment: segment, holder: holder, registry: r}, nil
}

func (r *SegmentLeaseRegistry) release(segment, holder string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.leases[segment]; !ok || h != holder {
		return
	}
	delete(r.leases, segment)
	delete(r.expires, segment)
}

func (r *SegmentLeaseRegistry) ReleaseHolder(segment, holder string) {
	r.release(segment, holder)
}

func (r *SegmentLeaseRegistry) IsHeld(segment string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.leases[segment]
	if !ok {
		return false
	}
	exp, ok := r.expires[segment]
	return ok && time.Now().Before(exp)
}
