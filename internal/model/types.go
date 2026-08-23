package model

import "time"

type TowerState string

const (
	TowerIdle      TowerState = "idle"
	TowerPriming   TowerState = "priming"
	TowerOperating TowerState = "operating"
	TowerDriftHold TowerState = "drift_hold"
	TowerFault     TowerState = "fault"
	TowerShutdown  TowerState = "shutdown"
)

type FanState string

const (
	FanOff     FanState = "off"
	FanStaging FanState = "staging"
	FanRun     FanState = "run"
	FanCoast   FanState = "coast"
	FanTrip    FanState = "trip"
)

type SprayState string

const (
	SprayClosed    SprayState = "closed"
	SprayPriming   SprayState = "priming"
	SprayActive    SprayState = "active"
	SprayThrottled SprayState = "throttled"
	SprayFault     SprayState = "fault"
)

type FlowSetpoint struct {
	GallonsPerMinute float64
	TolerancePct     float64
}

func (f FlowSetpoint) Within(actual float64) bool {
	if f.GallonsPerMinute <= 0 {
		return actual <= 0
	}
	lo := f.GallonsPerMinute * (1 - f.TolerancePct/100)
	hi := f.GallonsPerMinute * (1 + f.TolerancePct/100)
	return actual >= lo && actual <= hi
}

type DriftReading struct {
	Sensor DriftSensorID
	PPM    float64
	At     time.Time
}

type FanAssignment struct {
	Fan      FanID
	SpeedPct float64
	Enabled  bool
	State    FanState
}

type SprayAssignment struct {
	Header   SprayHeaderID
	Setpoint FlowSetpoint
	Enabled  bool
	LastFlow float64
	State    SprayState
}

type TowerSnapshot struct {
	ID        TowerID
	State     TowerState
	Fans      []FanAssignment
	Sprays    []SprayAssignment
	DriftPPM  float64
	BasinTemp float64
	UpdatedAt time.Time
}

type SprayScheduleEntry struct {
	ID          ScheduleID
	Header      SprayHeaderID
	Start       time.Time
	End         time.Time
	Setpoint    FlowSetpoint
	MaxDriftPPM float64
}

type SpraySchedule struct {
	ID      ScheduleID
	Entries []SprayScheduleEntry
	Version int64
}

func (s SpraySchedule) Clone() SpraySchedule {
	out := SpraySchedule{ID: s.ID, Version: s.Version}
	if len(s.Entries) == 0 {
		return out
	}
	out.Entries = make([]SprayScheduleEntry, len(s.Entries))
	copy(out.Entries, s.Entries)
	return out
}

type AlarmEvent struct {
	Code     AlarmCode
	Message  string
	Tower    TowerID
	RaisedAt time.Time
	Severity int
}

type DriftBudget struct {
	MaxPPM           float64
	FanSpeedFactor   float64
	SprayFlowFactor  float64
	HoldDuration     time.Duration
	AllowSprayDuring bool
}

func (b DriftBudget) Exceeded(reading float64, fanPct, sprayGPM float64) bool {
	allowed := b.MaxPPM - (fanPct/100)*b.FanSpeedFactor - sprayGPM*b.SprayFlowFactor
	return reading > allowed
}

type TelemetryFrame struct {
	TowerID   TowerID     `json:"tower_id"`
	DriftPPM  float64     `json:"drift_ppm"`
	FanSpeed  float64     `json:"fan_speed"`
	SprayGPM  float64     `json:"spray_gpm"`
	BasinTemp float64     `json:"basin_temp"`
	State     TowerState  `json:"state"`
	At        time.Time   `json:"at"`
}
