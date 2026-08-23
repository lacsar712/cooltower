package config

import "testing"

func TestDefaultValid(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejectsZeroFans(t *testing.T) {
	cfg := Default()
	cfg.FanCount = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("COOLTOWER_TOWER_ID", "tower-test")
	t.Setenv("COOLTOWER_FAN_COUNT", "6")
	cfg := LoadFromEnv()
	if cfg.TowerID != "tower-test" || cfg.FanCount != 6 {
		t.Fatalf("env load: %+v", cfg)
	}
}

func TestDriftBudgetConfig(t *testing.T) {
	cfg := Default()
	b := cfg.DriftBudget()
	if b.MaxPPM != cfg.DriftMaxPPM {
		t.Fatal("budget max ppm mismatch")
	}
}
