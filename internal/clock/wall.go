package clock

import "time"

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type WallClock struct{}

func NewWall() *WallClock { return &WallClock{} }

func (WallClock) Now() time.Time { return time.Now() }

func (WallClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
