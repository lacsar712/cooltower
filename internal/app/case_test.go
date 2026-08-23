package app

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/cooltower/internal/config"
)

func TestCase(t *testing.T) {
	a, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = a.FanRamp(ctx)
	}()
	time.Sleep(15 * time.Millisecond)
	cancel()
	<-done
	if a.Telemetry().FanSpeed > 10 {
		t.Fatalf("fan ramp continued after cancel, speed=%.1f", a.Telemetry().FanSpeed)
	}
}
