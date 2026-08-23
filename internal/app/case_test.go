package app

import (
	"context"
	"errors"
	"testing"

	"github.com/lacsar712/cooltower/internal/config"
	"github.com/lacsar712/cooltower/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	anchor := a.Clock().Now()
	err = a.ConfirmAntifreeze(context.Background(), anchor)
	if err == nil {
		t.Fatal("expected antifreeze hold error")
	}
	if !errors.Is(err, model.ErrAntifreezeHold) {
		t.Fatalf("expected ErrAntifreezeHold, got %v", err)
	}
}
