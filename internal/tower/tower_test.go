package tower

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/model"
)

func TestBasinFillAndDraw(t *testing.T) {
	b := NewBasin("basin-1", 50)
	start := time.Unix(0, 0)
	if err := b.Fill(context.Background(), start); err != nil {
		t.Fatal(err)
	}
	if b.Level() < 85 {
		t.Fatal("expected fill")
	}
	if err := b.Draw(10); err != nil {
		t.Fatal(err)
	}
}

func TestUnitPrime(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	basin := NewBasin("b1", 40)
	unit := NewUnit("tower-1", basin, clk)
	if err := unit.Prime(context.Background()); err != nil {
		t.Fatal(err)
	}
	unit.SetDriftReading(12.5)
	snap := unit.Snapshot(model.TowerOperating, nil, nil)
	if snap.DriftPPM != 12.5 {
		t.Fatal(snap.DriftPPM)
	}
}

func TestRegistry(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	reg := NewRegistry()
	u := NewUnit("t1", NewBasin("b1", 50), clk)
	reg.Register(u)
	got, ok := reg.Get("t1")
	if !ok || got.ID() != "t1" {
		t.Fatal("registry get")
	}
	if len(reg.List()) != 1 {
		t.Fatal("registry list")
	}
}

func TestSnapshotCopiesSlices(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	unit := NewUnit("t1", NewBasin("b1", 50), clk)
	fans := []model.FanAssignment{{Fan: "f1", SpeedPct: 80}}
	snap := unit.Snapshot(model.TowerOperating, fans, nil)
	fans[0].SpeedPct = 0
	if snap.Fans[0].SpeedPct != 80 {
		t.Fatal("snapshot should copy fan slice")
	}
}
