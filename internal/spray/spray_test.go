package spray

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/model"
)

func TestHeaderOpenClose(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	sp := model.FlowSetpoint{GallonsPerMinute: 45, TolerancePct: 10}
	hdr, err := NewHeader("hdr-a", clk, 4, sp)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := hdr.Open(ctx); err != nil {
		t.Fatal(err)
	}
	if hdr.State() != model.SprayActive {
		t.Fatal(hdr.State())
	}
	if hdr.Flow() <= 0 {
		t.Fatal("expected flow")
	}
	if err := hdr.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if hdr.Flow() != 0 {
		t.Fatal("expected zero flow after close")
	}
}

func TestPlantTotalFlow(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	sp := model.FlowSetpoint{GallonsPerMinute: 30, TolerancePct: 10}
	plant := NewPlant()
	h1, _ := NewHeader("h1", clk, 2, sp)
	h2, _ := NewHeader("h2", clk, 2, sp)
	plant.Add(h1)
	plant.Add(h2)
	ctx := context.Background()
	if err := plant.OpenAll(ctx); err != nil {
		t.Fatal(err)
	}
	if plant.TotalFlow() <= 0 {
		t.Fatal("total flow")
	}
}

func TestNozzleThrottle(t *testing.T) {
	nz := NewNozzle("nz-1")
	_ = nz.Open(context.Background())
	nz.Throttle(50)
	if nz.Flow() <= 0 {
		t.Fatal("throttle flow")
	}
	nz.Close()
	if nz.IsOpen() {
		t.Fatal("closed")
	}
}

func TestValidateFlows(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Millisecond)
	sp := model.FlowSetpoint{GallonsPerMinute: 1000, TolerancePct: 1}
	hdr, _ := NewHeader("h1", clk, 2, sp)
	plant := NewPlant()
	plant.Add(hdr)
	ctx := context.Background()
	_ = hdr.Open(ctx)
	if err := plant.ValidateFlows(ctx); err == nil {
		t.Fatal("expected setpoint violation")
	}
}
