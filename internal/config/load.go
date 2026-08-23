package config

import (
	"os"
	"strconv"
	"strings"
)

func LoadFromEnv() Config {
	cfg := Default()
	if v := os.Getenv("COOLTOWER_TOWER_ID"); v != "" {
		cfg.TowerID = v
	}
	if v := os.Getenv("COOLTOWER_FAN_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.FanCount = n
		}
	}
	if v := os.Getenv("COOLTOWER_SPRAY_GPM"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.DefaultSprayGPM = f
		}
	}
	if v := os.Getenv("COOLTOWER_DRIFT_MAX_PPM"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.DriftMaxPPM = f
		}
	}
	if v := os.Getenv("COOLTOWER_WEB_ADDR"); v != "" {
		cfg.WebListenAddr = strings.TrimSpace(v)
	}
	return cfg
}
