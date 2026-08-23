package drift

import (
	"sync"
	"time"

	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/model"
)

type Sensor struct {
	mu     sync.RWMutex
	id     model.DriftSensorID
	ppm    float64
	lastAt time.Time
}

func NewSensor(id model.DriftSensorID) *Sensor {
	return &Sensor{id: id}
}

func (s *Sensor) ID() model.DriftSensorID { return s.id }

func (s *Sensor) Observe(ppm float64, at time.Time) model.DriftReading {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ppm = ppm
	s.lastAt = at
	return model.DriftReading{Sensor: s.id, PPM: ppm, At: at}
}

func (s *Sensor) Latest() model.DriftReading {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return model.DriftReading{Sensor: s.id, PPM: s.ppm, At: s.lastAt}
}

type Window struct {
	mu       sync.RWMutex
	start    time.Time
	duration time.Duration
	maxPPM   float64
	clk      clock.Clock
}

func NewWindow(clk clock.Clock, start time.Time, duration time.Duration, maxPPM float64) *Window {
	return &Window{clk: clk, start: start, duration: duration, maxPPM: maxPPM}
}

func (w *Window) Active() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return clock.WindowElapsed(w.clk, w.start, w.duration)
}

func (w *Window) Closed() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return clock.WindowClosed(w.clk, w.start, w.duration)
}

func (w *Window) MaxPPM() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.maxPPM
}

func (w *Window) Violated(reading float64) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.Active() && reading > w.maxPPM
}

func (w *Window) Extend(d time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.duration += d
}

type Aggregator struct {
	mu      sync.RWMutex
	sensors []*Sensor
	clk     clock.Clock
}

func NewAggregator(clk clock.Clock) *Aggregator {
	return &Aggregator{clk: clk}
}

func (a *Aggregator) Add(s *Sensor) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sensors = append(a.sensors, s)
}

func (a *Aggregator) PeakPPM() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var peak float64
	for _, s := range a.sensors {
		r := s.Latest()
		if r.PPM > peak {
			peak = r.PPM
		}
	}
	return peak
}

func (a *Aggregator) AveragePPM() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.sensors) == 0 {
		return 0
	}
	var sum float64
	for _, s := range a.sensors {
		sum += s.Latest().PPM
	}
	return sum / float64(len(a.sensors))
}

func (a *Aggregator) Readings() []model.DriftReading {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]model.DriftReading, 0, len(a.sensors))
	for _, s := range a.sensors {
		out = append(out, s.Latest())
	}
	return out
}

func (a *Aggregator) ObserveAll(ppm float64, at time.Time) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, s := range a.sensors {
		s.Observe(ppm, at)
	}
}

func (a *Aggregator) SensorCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.sensors)
}
