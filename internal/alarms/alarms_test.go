package alarms

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
)

func TestEmitterRaise(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	em := NewEmitter(NewRegistry(), clk, 8)
	ctx := context.Background()
	if err := em.Raise(ctx, "DRIFT_HIGH", "tower-1", 2); err != nil {
		t.Fatal(err)
	}
	active := em.Active()
	if len(active) != 1 || active[0].Code != "DRIFT_HIGH" {
		t.Fatal(active)
	}
}

func TestEmitterClear(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	em := NewEmitter(NewRegistry(), clk, 8)
	ctx := context.Background()
	_ = em.Raise(ctx, "DRIFT_HIGH", "t1", 1)
	em.Clear("DRIFT_HIGH")
	if len(em.Active()) != 0 {
		t.Fatal("expected cleared")
	}
	if len(em.History()) != 1 {
		t.Fatal("history preserved")
	}
}

func TestRegistryMessage(t *testing.T) {
	reg := NewRegistry()
	if reg.Message("DRIFT_HIGH") == "" {
		t.Fatal("expected message")
	}
	reg.Register("CUSTOM", "custom alarm")
	if reg.Message("CUSTOM") != "custom alarm" {
		t.Fatal(reg.Message("CUSTOM"))
	}
}
