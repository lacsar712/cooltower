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
	err = a.ReportDriftFault(context.Background(), 100)
	if err == nil {
		t.Fatal("expected drift exceeded error")
	}
	if !errors.Is(err, model.ErrDriftExceeded) {
		t.Fatalf("expected ErrDriftExceeded, got %v", err)
	}
}
