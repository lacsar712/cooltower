package model

import (
	"fmt"
	"strings"
)

type TowerID string
type FanID string
type SprayHeaderID string
type NozzleID string
type DriftSensorID string
type BasinID string
type ScheduleID string
type AlarmCode string

func (id TowerID) String() string         { return string(id) }
func (id FanID) String() string           { return string(id) }
func (id SprayHeaderID) String() string   { return string(id) }
func (id NozzleID) String() string        { return string(id) }
func (id DriftSensorID) String() string   { return string(id) }
func (id BasinID) String() string         { return string(id) }
func (id ScheduleID) String() string     { return string(id) }
func (id AlarmCode) String() string       { return string(id) }

func ParseTowerID(raw string) (TowerID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return TowerID(raw), nil
}

func ParseFanID(raw string) (FanID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return FanID(raw), nil
}

func ParseSprayHeaderID(raw string) (SprayHeaderID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return SprayHeaderID(raw), nil
}

func ParseNozzleID(header SprayHeaderID, index int) (NozzleID, error) {
	if header == "" || index < 0 {
		return "", ErrInvalidID
	}
	return NozzleID(fmt.Sprintf("%s-nz-%02d", header, index)), nil
}

func ParseDriftSensorID(raw string) (DriftSensorID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return DriftSensorID(raw), nil
}

func ParseBasinID(raw string) (BasinID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return BasinID(raw), nil
}

func ParseScheduleID(raw string) (ScheduleID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidID
	}
	return ScheduleID(raw), nil
}
