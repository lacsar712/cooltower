package model

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidID       = errors.New("cooltower: invalid identifier")
	ErrNotFound        = errors.New("cooltower: entity not found")
	ErrConflict        = errors.New("cooltower: state conflict")
	ErrInterlock       = errors.New("cooltower: interlock denied")
	ErrDriftHold       = errors.New("cooltower: drift hold active")
	ErrFlowSetpoint    = errors.New("cooltower: flow setpoint violation")
	ErrFanFault        = errors.New("cooltower: fan fault")
	ErrSprayFault      = errors.New("cooltower: spray fault")
	ErrScheduleEmpty   = errors.New("cooltower: schedule empty")
	ErrContextCanceled = errors.New("cooltower: operation canceled")
)

type DomainError struct {
	Op   string
	Code string
	Err  error
}

func (e *DomainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("cooltower %s [%s]: %v", e.Op, e.Code, e.Err)
	}
	return fmt.Sprintf("cooltower %s [%s]", e.Op, e.Code)
}

func (e *DomainError) Unwrap() error { return e.Err }

func Wrap(op, code string, err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Op: op, Code: code, Err: err}
}

func Is(err, target error) bool     { return errors.Is(err, target) }
func As(err error, target any) bool { return errors.As(err, target) }
