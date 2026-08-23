package fan

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/model"
)

func TestBankStartStop(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	bank := NewBank("fan-1", clk, time.Millisecond, 5*time.Millisecond)
	ctx := context.Background()
	if err := bank.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if bank.State() != model.FanRun {
		t.Fatal(bank.State())
	}
	if err := bank.SetSpeed(75); err != nil {
		t.Fatal(err)
	}
	if bank.Speed() != 75 {
		t.Fatal(bank.Speed())
	}
	if err := bank.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if bank.State() != model.FanOff {
		t.Fatal(bank.State())
	}
}

func TestVFDTrip(t *testing.T) {
	v := NewVFD("f1", time.Second, time.Second)
	ctx := context.Background()
	if err := v.Enable(ctx); err != nil {
		t.Fatal(err)
	}
	if err := v.Trip(ctx); err != nil {
		t.Fatal(err)
	}
	if err := v.Enable(ctx); err == nil {
		t.Fatal("expected trip block")
	}
}

func TestCoordinatorAverage(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	coord := NewCoordinator()
	b1 := NewBank("f1", clk, time.Millisecond, time.Millisecond)
	b2 := NewBank("f2", clk, time.Millisecond, time.Millisecond)
	coord.Add(b1)
	coord.Add(b2)
	ctx := context.Background()
	if err := coord.StartAll(ctx); err != nil {
		t.Fatal(err)
	}
	_ = b1.SetSpeed(60)
	_ = b2.SetSpeed(80)
	avg := coord.AverageSpeed()
	if avg != 70 {
		t.Fatalf("avg %f", avg)
	}
}

func TestSetSpeedNotRunning(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	bank := NewBank("fan-1", clk, time.Millisecond, time.Millisecond)
	if err := bank.SetSpeed(50); err == nil {
		t.Fatal("expected not running error")
	}
}
