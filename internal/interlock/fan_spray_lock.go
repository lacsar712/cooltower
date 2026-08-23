package interlock

import (
	"context"
	"sync"
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

type lease struct {
	key   string
	until time.Time
}

type FanSprayLock struct {
	mu     sync.Mutex
	holder map[string]lease
	clk    func() time.Time
}

func NewFanSprayLock(now func() time.Time) *FanSprayLock {
	if now == nil {
		now = time.Now
	}
	return &FanSprayLock{holder: make(map[string]lease), clk: now}
}

func (l *FanSprayLock) keyFor(fan model.FanID, spray model.SprayHeaderID) string {
	return string(fan) + ":" + string(spray)
}

func (l *FanSprayLock) TryAcquire(fan model.FanID, spray model.SprayHeaderID, ttl time.Duration) (release func(), ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := l.keyFor(fan, spray)
	now := l.clk()
	if ex, exists := l.holder[key]; exists && now.Before(ex.until) {
		return nil, false
	}
	until := now.Add(ttl)
	l.holder[key] = lease{key: key, until: until}
	var once sync.Once
	release = func() {
		once.Do(func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			if cur, ok := l.holder[key]; ok && cur.until.Equal(until) {
				delete(l.holder, key)
			}
		})
	}
	return release, true
}

func (l *FanSprayLock) WithLease(ctx context.Context, fan model.FanID, spray model.SprayHeaderID, ttl time.Duration, fn func() error) error {
	release, ok := l.TryAcquire(fan, spray, ttl)
	if !ok {
		return model.Wrap("fan_spray_lock", "busy", model.ErrInterlock)
	}
	defer release()
	done := make(chan error, 1)
	go func() { done <- fn() }()
	select {
	case <-ctx.Done():
		return model.Wrap("fan_spray_lock", "canceled", context.Cause(ctx))
	case err := <-done:
		return err
	}
}

func (l *FanSprayLock) IsHeld(fan model.FanID, spray model.SprayHeaderID) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := l.keyFor(fan, spray)
	ex, ok := l.holder[key]
	if !ok {
		return false
	}
	return l.clk().Before(ex.until)
}
