package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/lacsar712/cooltower/internal/config"
	"github.com/lacsar712/cooltower/internal/model"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	fails := 0
	CalibrateProbe = func(ctx context.Context) error {
		fails++
		if fails == 1 {
			return fmt.Errorf("sensor fault")
		}
		return nil
	}
	defer func() { CalibrateProbe = nil }()
	ctx := context.Background()
	segment := model.SprayHeaderID(a.TowerID().String() + "-spray-01")
	if err := a.CalibrateSpray(ctx, segment, "crew-a"); err == nil {
		t.Fatal("expected calibration failure")
	}
	if err := a.CalibrateSpray(ctx, segment, "crew-b"); err != nil {
		t.Fatalf("second holder blocked by leaked segment lease: %v", err)
	}
}
