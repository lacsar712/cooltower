package drift

import (
	"testing"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
)

func TestSensorObserve(t *testing.T) {
	s := NewSensor("drift-1")
	at := time.Unix(0, 0)
	r := s.Observe(25.5, at)
	if r.PPM != 25.5 {
		t.Fatal(r.PPM)
	}
	if s.Latest().PPM != 25.5 {
		t.Fatal("latest")
	}
}

func TestWindowViolation(t *testing.T) {
	start := time.Unix(0, 0)
	clk := clock.NewProcessClock(start, time.Second)
	w := NewWindow(clk, start, time.Minute, 30)
	if !w.Active() {
		t.Fatal("window active")
	}
	if !w.Violated(40) {
		t.Fatal("expected violation")
	}
	clk.Advance(2 * time.Minute)
	if w.Active() {
		t.Fatal("window closed")
	}
}

func TestAggregatorPeak(t *testing.T) {
	clk := clock.NewProcessClock(time.Unix(0, 0), time.Second)
	agg := NewAggregator(clk)
	s1 := NewSensor("d1")
	s2 := NewSensor("d2")
	agg.Add(s1)
	agg.Add(s2)
	s1.Observe(20, time.Unix(0, 0))
	s2.Observe(35, time.Unix(0, 0))
	if agg.PeakPPM() != 35 {
		t.Fatal(agg.PeakPPM())
	}
	if agg.AveragePPM() != 27.5 {
		t.Fatal(agg.AveragePPM())
	}
}
