package fsm

import (
	"context"
	"testing"

	"github.com/lacsar712/cooltower/internal/model"
)

func TestTowerFSMHappyPath(t *testing.T) {
	fsm := NewTowerFSM("t1", nil)
	ctx := context.Background()
	for _, ev := range []string{"prime", "spray_ok"} {
		if err := fsm.Apply(ctx, ev); err != nil {
			t.Fatalf("event %s: %v", ev, err)
		}
	}
	if fsm.State() != model.TowerOperating {
		t.Fatalf("state %s", fsm.State())
	}
}

func TestTowerFSMIllegal(t *testing.T) {
	fsm := NewTowerFSM("t1", nil)
	if err := fsm.Apply(context.Background(), "drift_high"); err == nil {
		t.Fatal("expected illegal transition")
	}
}

func TestFanFSMStaging(t *testing.T) {
	fsm := NewFanFSM("f1", nil)
	ctx := context.Background()
	if err := fsm.Apply(ctx, "start"); err != nil {
		t.Fatal(err)
	}
	if err := fsm.Apply(ctx, "staged"); err != nil {
		t.Fatal(err)
	}
	if fsm.State() != model.FanRun {
		t.Fatal(fsm.State())
	}
}

func TestSprayFSMThrottle(t *testing.T) {
	fsm := NewSprayFSM("hdr-1", nil)
	ctx := context.Background()
	for _, ev := range []string{"open", "flow_ok", "throttle", "restore"} {
		if err := fsm.Apply(ctx, ev); err != nil {
			t.Fatalf("%s: %v", ev, err)
		}
	}
	if fsm.State() != model.SprayActive {
		t.Fatal(fsm.State())
	}
}

func TestTowerSideEffectRollback(t *testing.T) {
	fsm := NewTowerFSM("t1", func(ctx context.Context, tower model.TowerID, from, to model.TowerState) error {
		return model.ErrInterlock
	})
	if err := fsm.Apply(context.Background(), "prime"); err == nil {
		t.Fatal("expected side effect error")
	}
	if fsm.State() != model.TowerIdle {
		t.Fatal("state should not advance on side effect failure")
	}
}
