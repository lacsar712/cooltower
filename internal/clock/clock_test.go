package clock

import (
	"context"
	"testing"
	"time"
)

func TestProcessClockAdvance(t *testing.T) {
	start := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	clk := NewProcessClock(start, time.Millisecond)
	if !clk.Now().Equal(start) {
		t.Fatal("start time")
	}
	next := clk.Advance(time.Minute)
	if next.Sub(start) != time.Minute {
		t.Fatal("advance")
	}
}

func TestWindowElapsed(t *testing.T) {
	start := time.Unix(0, 0)
	clk := NewProcessClock(start, time.Second)
	if !WindowElapsed(clk, start, time.Minute) {
		t.Fatal("window should be open at start")
	}
	clk.Advance(2 * time.Minute)
	if WindowElapsed(clk, start, time.Minute) {
		t.Fatal("window should be closed")
	}
}

func TestWaitUntilContextCancel(t *testing.T) {
	clk := NewProcessClock(time.Unix(0, 0), time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if WaitUntilContext(ctx, clk, time.Unix(100, 0)) == nil {
		t.Fatal("expected cancel")
	}
}

func TestWallClock(t *testing.T) {
	w := NewWall()
	if w.Now().IsZero() {
		t.Fatal("wall now")
	}
}
