package interlock

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/model"
)

func TestFanSprayLeaseRelease(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewFanSprayLock(func() time.Time { return now })
	release, ok := l.TryAcquire("fan-1", "hdr-a", time.Second)
	if !ok {
		t.Fatal("acquire")
	}
	release()
	if _, ok := l.TryAcquire("fan-1", "hdr-a", time.Second); !ok {
		t.Fatal("reacquire")
	}
}

func TestGuardPermit(t *testing.T) {
	g := NewGuard(map[model.FanID]model.SprayHeaderID{"fan-1": "hdr-a"})
	if err := g.Permit("fan-1", "hdr-b"); err == nil {
		t.Fatal("expected mismatch")
	}
	if err := g.Permit("fan-1", "hdr-a"); err != nil {
		t.Fatal(err)
	}
}

func TestDriftGuardHold(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Second)
	budget := model.DriftBudget{
		MaxPPM: 30, FanSpeedFactor: 10, SprayFlowFactor: 0.2, HoldDuration: time.Minute,
	}
	guard := NewDriftGuard(clk, budget)
	ctx := context.Background()
	if err := guard.PermitSpray(ctx, 50, 80, 20); err == nil {
		t.Fatal("expected interlock")
	}
	if !guard.HoldActive() {
		t.Fatal("hold should be active")
	}
	clk.Advance(2 * time.Minute)
	if guard.HoldActive() {
		t.Fatal("hold should expire")
	}
}

func TestWithLeaseCancel(t *testing.T) {
	l := NewFanSprayLock(time.Now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if l.WithLease(ctx, "f1", "h1", time.Second, func() error { return nil }) == nil {
		t.Fatal("expected cancel")
	}
}

func TestDriftGuardFanSpeed(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Second)
	budget := model.DriftBudget{MaxPPM: 50, FanSpeedFactor: 20, SprayFlowFactor: 0.1}
	guard := NewDriftGuard(clk, budget)
	if err := guard.PermitFanSpeed(context.Background(), 60, 90, 5); err == nil {
		t.Fatal("expected fan speed interlock")
	}
}
