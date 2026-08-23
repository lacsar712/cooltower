package fsm

import (
	"fmt"

	"github.com/lacsar712/cooltower/internal/model"
)

type Transition struct {
	From  model.TowerState
	To    model.TowerState
	Event string
}

var towerTransitions = []Transition{
	{model.TowerIdle, model.TowerPriming, "prime"},
	{model.TowerPriming, model.TowerOperating, "spray_ok"},
	{model.TowerOperating, model.TowerDriftHold, "drift_high"},
	{model.TowerDriftHold, model.TowerOperating, "drift_clear"},
	{model.TowerOperating, model.TowerIdle, "stop"},
	{model.TowerPriming, model.TowerFault, "fault"},
	{model.TowerOperating, model.TowerFault, "fault"},
	{model.TowerDriftHold, model.TowerFault, "fault"},
	{model.TowerFault, model.TowerShutdown, "shutdown"},
	{model.TowerIdle, model.TowerShutdown, "shutdown"},
}

func AllowedTower(from model.TowerState, event string) (model.TowerState, bool) {
	for _, tr := range towerTransitions {
		if tr.From == from && tr.Event == event {
			return tr.To, true
		}
	}
	return from, false
}

func MustTower(from model.TowerState, event string) (model.TowerState, error) {
	to, ok := AllowedTower(from, event)
	if !ok {
		return from, model.Wrap("tower_fsm", "illegal_transition", fmt.Errorf("%s -> %s", from, event))
	}
	return to, nil
}

var fanTransitions = []struct {
	from, to model.FanState
	event    string
}{
	{model.FanOff, model.FanStaging, "start"},
	{model.FanStaging, model.FanRun, "staged"},
	{model.FanRun, model.FanCoast, "stop"},
	{model.FanCoast, model.FanOff, "coast_done"},
	{model.FanRun, model.FanTrip, "trip"},
	{model.FanStaging, model.FanTrip, "trip"},
}

func AllowedFan(from model.FanState, event string) (model.FanState, bool) {
	for _, tr := range fanTransitions {
		if tr.from == from && tr.event == event {
			return tr.to, true
		}
	}
	return from, false
}

func MustFan(from model.FanState, event string) (model.FanState, error) {
	to, ok := AllowedFan(from, event)
	if !ok {
		return from, model.Wrap("fan_fsm", "illegal_transition", fmt.Errorf("%s -> %s", from, event))
	}
	return to, nil
}

var sprayTransitions = []struct {
	from, to model.SprayState
	event    string
}{
	{model.SprayClosed, model.SprayPriming, "open"},
	{model.SprayPriming, model.SprayActive, "flow_ok"},
	{model.SprayActive, model.SprayThrottled, "throttle"},
	{model.SprayThrottled, model.SprayActive, "restore"},
	{model.SprayActive, model.SprayClosed, "close"},
	{model.SprayPriming, model.SprayFault, "fault"},
	{model.SprayActive, model.SprayFault, "fault"},
}

func AllowedSpray(from model.SprayState, event string) (model.SprayState, bool) {
	for _, tr := range sprayTransitions {
		if tr.from == from && tr.event == event {
			return tr.to, true
		}
	}
	return from, false
}

func MustSpray(from model.SprayState, event string) (model.SprayState, error) {
	to, ok := AllowedSpray(from, event)
	if !ok {
		return from, model.Wrap("spray_fsm", "illegal_transition", fmt.Errorf("%s -> %s", from, event))
	}
	return to, nil
}
