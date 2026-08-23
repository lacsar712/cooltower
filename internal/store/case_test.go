package store

import (
	"testing"
	"time"

	"github.com/lacsar712/cooltower/internal/model"
)

func TestCase(t *testing.T) {
	orig := DutySnapshot{
		TowerID: "tower-a1",
		SpraySlots: []model.SprayScheduleEntry{{
			Start: time.Unix(0, 0), End: time.Unix(100, 0),
			Setpoint: model.FlowSetpoint{GallonsPerMinute: 45, TolerancePct: 5},
		}},
	}
	clone := orig.Clone()
	clone.SpraySlots[0].Setpoint.GallonsPerMinute = 99
	if orig.SpraySlots[0].Setpoint.GallonsPerMinute == 99 {
		t.Fatal("clone mutated original SpraySlots backing array")
	}
}
