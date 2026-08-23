package model

import "testing"

func TestFlowSetpointWithin(t *testing.T) {
	sp := FlowSetpoint{GallonsPerMinute: 100, TolerancePct: 10}
	if !sp.Within(105) {
		t.Fatal("expected within tolerance")
	}
	if sp.Within(120) {
		t.Fatal("expected outside tolerance")
	}
}

func TestDriftBudgetExceeded(t *testing.T) {
	b := DriftBudget{MaxPPM: 50, FanSpeedFactor: 20, SprayFlowFactor: 0.5}
	if !b.Exceeded(60, 80, 10) {
		t.Fatal("expected drift exceeded")
	}
	if b.Exceeded(10, 20, 5) {
		t.Fatal("expected within budget")
	}
}

func TestSprayScheduleClone(t *testing.T) {
	s := SpraySchedule{
		ID: "sched-1",
		Entries: []SprayScheduleEntry{
			{ID: "e1", Header: "hdr-a", MaxDriftPPM: 30},
		},
		Version: 1,
	}
	clone := s.Clone()
	clone.Entries[0].MaxDriftPPM = 99
	if s.Entries[0].MaxDriftPPM != 30 {
		t.Fatal("clone should not alias entries")
	}
}

func TestParseIDs(t *testing.T) {
	if _, err := ParseTowerID(""); err == nil {
		t.Fatal("empty tower id")
	}
	tid, err := ParseTowerID("tower-1")
	if err != nil || tid != "tower-1" {
		t.Fatalf("parse tower: %v", err)
	}
	nz, err := ParseNozzleID("hdr-a", 3)
	if err != nil || nz != "hdr-a-nz-03" {
		t.Fatalf("parse nozzle: %v %s", err, nz)
	}
}
