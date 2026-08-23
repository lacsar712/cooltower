package interlock

import "github.com/lacsar712/cooltower/internal/model"

type FillGuard struct {
	threshold float64
}

func NewFillGuard(threshold float64) *FillGuard {
	return &FillGuard{threshold: threshold}
}

func (g *FillGuard) Permit(delta float64) error {
	if delta > g.threshold {
		return model.Wrap("fill_guard", "block", model.ErrFillBlock)
	}
	return nil
}
