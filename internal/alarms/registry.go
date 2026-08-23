package alarms

import (
	"sync"

	"github.com/lacsar712/cooltower/internal/model"
)

type Registry struct {
	mu    sync.RWMutex
	codes map[model.AlarmCode]string
}

func NewRegistry() *Registry {
	r := &Registry{codes: make(map[model.AlarmCode]string)}
	r.codes["DRIFT_HIGH"] = "Drift eliminator reading exceeded budget"
	r.codes["FAN_TRIP"] = "Fan bank tripped on overload"
	r.codes["SPRAY_FAULT"] = "Spray header fault detected"
	r.codes["BASIN_LOW"] = "Basin water level critically low"
	r.codes["INTERLOCK"] = "Fan/spray interlock denied operation"
	return r
}

func (r *Registry) Message(code model.AlarmCode) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if msg, ok := r.codes[code]; ok {
		return msg
	}
	return string(code)
}

func (r *Registry) Register(code model.AlarmCode, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes[code] = message
}

func (r *Registry) Codes() []model.AlarmCode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.AlarmCode, 0, len(r.codes))
	for c := range r.codes {
		out = append(out, c)
	}
	return out
}
