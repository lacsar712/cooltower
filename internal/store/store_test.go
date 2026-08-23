package store

import (
	"testing"
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

func TestMemoryTowerRoundTrip(t *testing.T) {
	mem := NewMemory()
	snap := NewSnapshotBuilder("t1").State(model.TowerOperating).Drift(12).Build(time.Unix(0, 0))
	mem.PutTower(snap)
	got, ok := mem.GetTower("t1")
	if !ok || got.DriftPPM != 12 {
		t.Fatal("round trip")
	}
}

func TestCloneSnapshot(t *testing.T) {
	orig := model.TowerSnapshot{
		ID: "t1", Sprays: []model.SprayAssignment{{Header: "h1", LastFlow: 10}},
	}
	clone := CloneSnapshot(orig)
	clone.Sprays[0].LastFlow = 99
	if orig.Sprays[0].LastFlow != 10 {
		t.Fatal("clone should not alias sprays")
	}
}

func TestScheduleActiveEntry(t *testing.T) {
	mem := NewMemory()
	ss := NewScheduleStore(mem)
	start := time.Unix(100, 0)
	sched := model.SpraySchedule{
		ID: "s1",
		Entries: []model.SprayScheduleEntry{
			{ID: "e1", Header: "h1", Start: start, End: start.Add(time.Hour), MaxDriftPPM: 25},
		},
	}
	ss.Save(sched)
	snap, err := ss.SnapshotClone("s1")
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := ss.ActiveEntry(snap, start.Add(30*time.Minute))
	if !ok || entry.Header != "h1" {
		t.Fatal("active entry")
	}
}

func TestDiffSprays(t *testing.T) {
	before := model.TowerSnapshot{
		Sprays: []model.SprayAssignment{{Header: "h1", LastFlow: 10, Enabled: true}},
	}
	after := model.TowerSnapshot{
		Sprays: []model.SprayAssignment{{Header: "h1", LastFlow: 20, Enabled: true}},
	}
	changed := DiffSprays(before, after)
	if len(changed) != 1 || changed[0] != "h1" {
		t.Fatal(changed)
	}
}
