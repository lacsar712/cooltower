package fsm

import (
	"context"

	"github.com/lacsar712/cooltower/internal/model"
)

type FanSideEffect func(ctx context.Context, fan model.FanID, from, to model.FanState) error

type FanFSM struct {
	id       model.FanID
	state    model.FanState
	onChange FanSideEffect
}

func NewFanFSM(id model.FanID, effect FanSideEffect) *FanFSM {
	return &FanFSM{id: id, state: model.FanOff, onChange: effect}
}

func (f *FanFSM) State() model.FanState { return f.state }

func (f *FanFSM) Apply(ctx context.Context, event string) error {
	next, err := MustFan(f.state, event)
	if err != nil {
		return err
	}
	prev := f.state
	if f.onChange != nil {
		if err := f.onChange(ctx, f.id, prev, next); err != nil {
			return model.Wrap("fan_fsm", "side_effect", err)
		}
	}
	f.state = next
	return nil
}

type SpraySideEffect func(ctx context.Context, header model.SprayHeaderID, from, to model.SprayState) error

type SprayFSM struct {
	id       model.SprayHeaderID
	state    model.SprayState
	onChange SpraySideEffect
}

func NewSprayFSM(id model.SprayHeaderID, effect SpraySideEffect) *SprayFSM {
	return &SprayFSM{id: id, state: model.SprayClosed, onChange: effect}
}

func (f *SprayFSM) State() model.SprayState { return f.state }

func (f *SprayFSM) Apply(ctx context.Context, event string) error {
	next, err := MustSpray(f.state, event)
	if err != nil {
		return err
	}
	prev := f.state
	if f.onChange != nil {
		if err := f.onChange(ctx, f.id, prev, next); err != nil {
			return model.Wrap("spray_fsm", "side_effect", err)
		}
	}
	f.state = next
	return nil
}
