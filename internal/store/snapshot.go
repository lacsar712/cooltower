package store

import (
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

type SnapshotBuilder struct {
	id     model.TowerID
	state  model.TowerState
	fans   []model.FanAssignment
	sprays []model.SprayAssignment
	drift  float64
	basin  float64
}

func NewSnapshotBuilder(id model.TowerID) *SnapshotBuilder {
	return &SnapshotBuilder{id: id, state: model.TowerIdle}
}

func (b *SnapshotBuilder) State(s model.TowerState) *SnapshotBuilder {
	b.state = s
	return b
}

func (b *SnapshotBuilder) Fan(a model.FanAssignment) *SnapshotBuilder {
	b.fans = append(b.fans, a)
	return b
}

func (b *SnapshotBuilder) Spray(a model.SprayAssignment) *SnapshotBuilder {
	b.sprays = append(b.sprays, a)
	return b
}

func (b *SnapshotBuilder) Drift(ppm float64) *SnapshotBuilder {
	b.drift = ppm
	return b
}

func (b *SnapshotBuilder) BasinTemp(c float64) *SnapshotBuilder {
	b.basin = c
	return b
}

func (b *SnapshotBuilder) Build(at time.Time) model.TowerSnapshot {
	fans := make([]model.FanAssignment, len(b.fans))
	copy(fans, b.fans)
	sprays := make([]model.SprayAssignment, len(b.sprays))
	copy(sprays, b.sprays)
	return model.TowerSnapshot{
		ID: b.id, State: b.state, Fans: fans, Sprays: sprays,
		DriftPPM: b.drift, BasinTemp: b.basin, UpdatedAt: at,
	}
}

func DiffSprays(before, after model.TowerSnapshot) []model.SprayHeaderID {
	index := make(map[model.SprayHeaderID]model.SprayAssignment)
	for _, s := range before.Sprays {
		index[s.Header] = s
	}
	var changed []model.SprayHeaderID
	for _, s := range after.Sprays {
		prev, ok := index[s.Header]
		if !ok || prev.LastFlow != s.LastFlow || prev.Enabled != s.Enabled || prev.State != s.State {
			changed = append(changed, s.Header)
		}
	}
	return changed
}

func CloneSnapshot(s model.TowerSnapshot) model.TowerSnapshot {
	fans := make([]model.FanAssignment, len(s.Fans))
	copy(fans, s.Fans)
	sprays := make([]model.SprayAssignment, len(s.Sprays))
	copy(sprays, s.Sprays)
	return model.TowerSnapshot{
		ID: s.ID, State: s.State, Fans: fans, Sprays: sprays,
		DriftPPM: s.DriftPPM, BasinTemp: s.BasinTemp, UpdatedAt: s.UpdatedAt,
	}
}

type DutySnapshot struct {
	TowerID    model.TowerID
	SpraySlots []model.SprayScheduleEntry
}

func (d DutySnapshot) Clone() DutySnapshot {
	out := DutySnapshot{TowerID: d.TowerID}
	slots := make([]model.SprayScheduleEntry, len(d.SpraySlots))
	copy(slots, d.SpraySlots)
	out.SpraySlots = slots
	return out
}
