package fsm

import (
	"context"
	"testing"

	"github.com/lacsar712/cooltower/internal/model"
)

func TestCase(t *testing.T) {
	FanDrivePulse = nil
	var pulses int
	FanDrivePulse = func() { pulses++ }
	defer func() { FanDrivePulse = nil }()
	f := NewTowerFSM(model.TowerID("tower-test"), nil)
	if err := f.Apply(context.Background(), "spray_ok"); err == nil {
		t.Fatal("expected illegal transition error")
	}
	if pulses != 0 {
		t.Fatalf("illegal transition should not pulse fan drive, got %d", pulses)
	}
}
