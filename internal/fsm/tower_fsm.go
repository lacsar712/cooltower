package fsm

import (
	"context"

	"github.com/lacsar712/cooltower/internal/model"
)

// FanDrivePulse counts fan-side effects from tower FSM transitions (acceptance).
var FanDrivePulse func()

type TowerSideEffect func(ctx context.Context, tower model.TowerID, from, to model.TowerState) error

type TowerFSM struct {
	id       model.TowerID
	state    model.TowerState
	onChange TowerSideEffect
}

func NewTowerFSM(id model.TowerID, effect TowerSideEffect) *TowerFSM {
	return &TowerFSM{id: id, state: model.TowerIdle, onChange: effect}
}

func (f *TowerFSM) State() model.TowerState { return f.state }

func (f *TowerFSM) Apply(ctx context.Context, event string) error {
	next, err := MustTower(f.state, event)
	if err != nil {
		return err
	}
	prev := f.state
	if f.onChange != nil {
		if err := f.onChange(ctx, f.id, prev, next); err != nil {
			return model.Wrap("tower_fsm", "side_effect", err)
		}
	}
	if next == model.TowerOperating && FanDrivePulse != nil {
		FanDrivePulse()
	}
	f.state = next
	return nil
}

func (f *TowerFSM) ForceState(s model.TowerState) {
	f.state = s
}
