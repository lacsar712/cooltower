package store

import (
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

type ScheduleStore struct {
	mem *Memory
}

func NewScheduleStore(mem *Memory) *ScheduleStore {
	return &ScheduleStore{mem: mem}
}

func (s *ScheduleStore) Save(sched model.SpraySchedule) {
	s.mem.PutSchedule(sched)
}

func (s *ScheduleStore) SnapshotClone(id model.ScheduleID) (model.SpraySchedule, error) {
	raw, ok := s.mem.GetSchedule(id)
	if !ok {
		return model.SpraySchedule{}, model.Wrap("schedule_store", "not_found", model.ErrNotFound)
	}
	return raw.Clone(), nil
}

func (s *ScheduleStore) ActiveEntry(snap model.SpraySchedule, now time.Time) (model.SprayScheduleEntry, bool) {
	for _, e := range snap.Entries {
		if !now.Before(e.Start) && now.Before(e.End) {
			return e, true
		}
	}
	return model.SprayScheduleEntry{}, false
}

func (s *ScheduleStore) EntriesInRange(snap model.SpraySchedule, from, to time.Time) []model.SprayScheduleEntry {
	var out []model.SprayScheduleEntry
	for _, e := range snap.Entries {
		if e.End.After(from) && e.Start.Before(to) {
			out = append(out, e)
		}
	}
	return out
}
