package clock

import (
	"context"
	"sync"

	"github.com/lacsar712/cooltower/internal/model"
)

type SprayScheduler struct {
	mu    sync.Mutex
	items []string
	clk   Clock
}

func NewSprayScheduler(clk Clock) *SprayScheduler {
	return &SprayScheduler{clk: clk}
}

func (s *SprayScheduler) ItemCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

func (s *SprayScheduler) InstallSprayPlanCtx(ctx context.Context, entries []model.SprayScheduleEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.mu.Lock()
		s.items = append(s.items, string(e.Header)+":"+e.Start.String())
		s.mu.Unlock()
	}
	return nil
}
