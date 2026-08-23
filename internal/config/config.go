package config

import "time"

type Config struct {
	TowerID            string
	FanCount           int
	SprayHeaderCount   int
	NozzlesPerHeader   int
	DefaultSprayGPM    float64
	FlowTolerancePct   float64
	DriftHoldMinutes   int
	DriftMaxPPM        float64
	FanMinRun          time.Duration
	FanCoast           time.Duration
	SprayPrimeSec      int
	AlarmBufferSize    int
	ProcessTickMs      int
	WebListenAddr      string
}

func Default() Config {
	return Config{
		TowerID: "tower-a1", FanCount: 4, SprayHeaderCount: 2, NozzlesPerHeader: 12,
		DefaultSprayGPM: 45.0, FlowTolerancePct: 5, DriftHoldMinutes: 2,
		DriftMaxPPM: 35.0, FanMinRun: time.Second, FanCoast: 3 * time.Second,
		SprayPrimeSec: 8, AlarmBufferSize: 64, ProcessTickMs: 10, WebListenAddr: ":8080",
	}
}

func (c Config) Validate() error {
	if c.FanCount <= 0 {
		return errConfig("fan_count must be positive")
	}
	if c.SprayHeaderCount <= 0 {
		return errConfig("spray_header_count must be positive")
	}
	if c.DefaultSprayGPM < 0 {
		return errConfig("default_spray_gpm invalid")
	}
	if c.DriftMaxPPM <= 0 {
		return errConfig("drift_max_ppm must be positive")
	}
	return nil
}

func (c Config) ProcessTick() time.Duration {
	if c.ProcessTickMs <= 0 {
		return 100 * time.Millisecond
	}
	return time.Duration(c.ProcessTickMs) * time.Millisecond
}

func (c Config) DriftBudget() DriftBudgetConfig {
	return DriftBudgetConfig{
		MaxPPM:          c.DriftMaxPPM,
		FanSpeedFactor:  c.DriftMaxPPM * 0.4,
		SprayFlowFactor: 0.15,
	}
}

type DriftBudgetConfig struct {
	MaxPPM          float64
	FanSpeedFactor  float64
	SprayFlowFactor float64
}

type errConfig string

func (e errConfig) Error() string { return "cooltower config: " + string(e) }
