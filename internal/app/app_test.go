package app

import (
	"context"
	"testing"

	"github.com/lacsar712/cooltower/internal/config"
	"github.com/lacsar712/cooltower/internal/model"
)

func TestRunOnceHappyPath(t *testing.T) {
	cfg := config.Default()
	cfg.FanCount = 2
	cfg.SprayHeaderCount = 2
	cfg.DriftHoldMinutes = 0
	app, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if app.towerFSM.State() != model.TowerOperating {
		t.Fatalf("state %s", app.towerFSM.State())
	}
}

func TestSnapshotClone(t *testing.T) {
	cfg := config.Default()
	app, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	snap := app.Snapshot()
	if snap.ID != model.TowerID(cfg.TowerID) {
		t.Fatal(snap.ID)
	}
}

func TestTelemetryFrame(t *testing.T) {
	cfg := config.Default()
	app, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	tel := app.Telemetry()
	if tel.TowerID != model.TowerID(cfg.TowerID) {
		t.Fatal(tel.TowerID)
	}
}

func TestApplyScheduleEmpty(t *testing.T) {
	cfg := config.Default()
	app, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.ApplyScheduleSnapshot(context.Background(), "missing"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestStatusLine(t *testing.T) {
	cfg := config.Default()
	app, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	line := app.StatusLine()
	if line == "" {
		t.Fatal("empty status")
	}
}
