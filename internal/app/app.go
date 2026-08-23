package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/cooltower/internal/alarms"
	"github.com/lacsar712/cooltower/internal/clock"
	"github.com/lacsar712/cooltower/internal/config"
	"github.com/lacsar712/cooltower/internal/drift"
	"github.com/lacsar712/cooltower/internal/fan"
	"github.com/lacsar712/cooltower/internal/fsm"
	"github.com/lacsar712/cooltower/internal/interlock"
	"github.com/lacsar712/cooltower/internal/model"
	"github.com/lacsar712/cooltower/internal/spray"
	"github.com/lacsar712/cooltower/internal/store"
	"github.com/lacsar712/cooltower/internal/tower"
)

type App struct {
	cfg        config.Config
	clk        clock.Clock
	mem        *store.Memory
	sched      *store.ScheduleStore
	unit       *tower.Unit
	towerFSM   *fsm.TowerFSM
	fanCoord   *fan.Coordinator
	sprayPlant *spray.Plant
	driftAgg   *drift.Aggregator
	driftWin   *drift.Window
	alarms     *alarms.Emitter
	driftGuard *interlock.DriftGuard
	guard      *interlock.Guard
	lock       *interlock.FanSprayLock
}

func New(cfg config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	clk := clock.NewProcessClock(start, cfg.ProcessTick())
	mem := store.NewMemory()
	towerID, err := model.ParseTowerID(cfg.TowerID)
	if err != nil {
		return nil, err
	}
	basinID, err := model.ParseBasinID(towerID.String() + "-basin")
	if err != nil {
		return nil, err
	}
	basin := tower.NewBasin(basinID, 60)
	unit := tower.NewUnit(towerID, basin, clk)
	a := &App{
		cfg: cfg, clk: clk, mem: mem, sched: store.NewScheduleStore(mem),
		unit: unit, fanCoord: fan.NewCoordinator(), sprayPlant: spray.NewPlant(),
		driftAgg: drift.NewAggregator(clk),
		lock: interlock.NewFanSprayLock(clk.Now),
	}
	budgetCfg := cfg.DriftBudget()
	a.driftGuard = interlock.NewDriftGuard(clk, model.DriftBudget{
		MaxPPM: budgetCfg.MaxPPM, FanSpeedFactor: budgetCfg.FanSpeedFactor,
		SprayFlowFactor: budgetCfg.SprayFlowFactor,
		HoldDuration: time.Duration(cfg.DriftHoldMinutes) * time.Minute,
	})
	a.alarms = alarms.NewEmitter(alarms.NewRegistry(), clk, cfg.AlarmBufferSize)
	a.towerFSM = fsm.NewTowerFSM(towerID, a.onTowerTransition)
	if err := a.wireFansAndSprays(towerID); err != nil {
		return nil, err
	}
	a.wireDriftSensors()
	a.driftWin = drift.NewWindow(clk, start, time.Duration(cfg.DriftHoldMinutes)*time.Minute, cfg.DriftMaxPPM)
	a.persistSnapshot(towerID)
	return a, nil
}

func (a *App) wireFansAndSprays(towerID model.TowerID) error {
	pairs := make(map[model.FanID]model.SprayHeaderID)
	sp := model.FlowSetpoint{GallonsPerMinute: a.cfg.DefaultSprayGPM, TolerancePct: a.cfg.FlowTolerancePct}
	for i := 0; i < a.cfg.FanCount; i++ {
		fanID, err := model.ParseFanID(fmt.Sprintf("%s-fan-%02d", towerID, i+1))
		if err != nil {
			return err
		}
		bank := fan.NewBank(fanID, a.clk, a.cfg.FanMinRun, a.cfg.FanCoast)
		a.fanCoord.Add(bank)
	}
	for i := 0; i < a.cfg.SprayHeaderCount; i++ {
		hdrID, err := model.ParseSprayHeaderID(fmt.Sprintf("%s-spray-%02d", towerID, i+1))
		if err != nil {
			return err
		}
		hdr, err := spray.NewHeader(hdrID, a.clk, a.cfg.NozzlesPerHeader, sp)
		if err != nil {
			return err
		}
		a.sprayPlant.Add(hdr)
	}
	fans := a.fanCoord.Assignments()
	sprays := a.sprayPlant.Assignments()
	for j := 0; j < len(fans) && j < len(sprays); j++ {
		pairs[fans[j].Fan] = sprays[j].Header
	}
	a.guard = interlock.NewGuard(pairs)
	return nil
}

func (a *App) wireDriftSensors() {
	for i := 0; i < 2; i++ {
		id, _ := model.ParseDriftSensorID(fmt.Sprintf("%s-drift-%d", a.cfg.TowerID, i+1))
		s := drift.NewSensor(id)
		a.driftAgg.Add(s)
	}
}

func (a *App) onTowerTransition(ctx context.Context, tower model.TowerID, from, to model.TowerState) error {
	switch to {
	case model.TowerFault:
		return a.alarms.Raise(ctx, "FAN_TRIP", tower, 3)
	case model.TowerDriftHold:
		return a.alarms.Raise(ctx, "DRIFT_HIGH", tower, 2)
	case model.TowerOperating:
		a.alarms.Clear("DRIFT_HIGH")
	}
	return nil
}

func (a *App) persistSnapshot(id model.TowerID) {
	b := store.NewSnapshotBuilder(id).State(a.towerFSM.State()).Drift(a.driftAgg.PeakPPM()).BasinTemp(a.unit.BasinTemperature())
	for _, f := range a.fanCoord.Assignments() {
		b.Fan(f)
	}
	for _, s := range a.sprayPlant.Assignments() {
		b.Spray(s)
	}
	a.mem.PutTower(b.Build(a.clk.Now()))
}

func (a *App) ApplyScheduleSnapshot(ctx context.Context, id model.ScheduleID) error {
	snap, err := a.sched.SnapshotClone(id)
	if err != nil {
		return err
	}
	now := a.clk.Now()
	entry, ok := a.sched.ActiveEntry(snap, now)
	if !ok {
		return model.Wrap("app", "schedule", model.ErrScheduleEmpty)
	}
	if hdr, ok := a.sprayPlant.Get(entry.Header); ok {
		hdr.BindSetpoint(entry.Setpoint)
	}
	a.driftWin = drift.NewWindow(a.clk, now, entry.End.Sub(entry.Start), entry.MaxDriftPPM)
	return nil
}

func (a *App) checkDriftInterlock(ctx context.Context) error {
	driftPPM := a.driftAgg.PeakPPM()
	fanSpeed := a.fanCoord.AverageSpeed()
	sprayGPM := a.sprayPlant.TotalFlow()
	if err := a.driftGuard.PermitSpray(ctx, driftPPM, fanSpeed, sprayGPM); err != nil {
		if model.Is(err, model.ErrInterlock) || model.Is(err, model.ErrDriftHold) {
			_ = a.towerFSM.Apply(ctx, "drift_high")
		}
		return err
	}
	if a.towerFSM.State() == model.TowerDriftHold {
		a.driftGuard.ClearHold()
		return a.towerFSM.Apply(ctx, "drift_clear")
	}
	return nil
}

func (a *App) RunOnce(ctx context.Context) error {
	if err := a.towerFSM.Apply(ctx, "prime"); err != nil {
		return err
	}
	if err := a.unit.Prime(ctx); err != nil {
		return err
	}
	if err := a.fanCoord.StartAll(ctx); err != nil {
		return err
	}
	fans := a.fanCoord.Assignments()
	for i := range fans {
		bank, ok := a.fanCoord.Bank(fans[i].Fan)
		if !ok {
			continue
		}
		sprayHdr, ok := a.guard.SpraysFor(fans[i].Fan)
		if !ok {
			continue
		}
		if err := a.guard.Permit(fans[i].Fan, sprayHdr); err != nil {
			return err
		}
		err := a.lock.WithLease(ctx, fans[i].Fan, sprayHdr, 30*time.Second, func() error {
			return bank.SetSpeed(70)
		})
		if err != nil {
			return err
		}
	}
	if err := a.sprayPlant.OpenAll(ctx); err != nil {
		return err
	}
	if err := a.towerFSM.Apply(ctx, "spray_ok"); err != nil {
		return err
	}
	a.observeDrift(15)
	if err := a.sprayPlant.ValidateFlows(ctx); err != nil {
		return err
	}
	if err := a.checkDriftInterlock(ctx); err != nil && !model.Is(err, model.ErrDriftHold) {
		if !model.Is(err, model.ErrInterlock) {
			return err
		}
	}
	if pc, ok := a.clk.(*clock.ProcessClock); ok {
		pc.Advance(time.Duration(a.cfg.DriftHoldMinutes)*time.Minute + time.Second)
	}
	if a.towerFSM.State() == model.TowerDriftHold {
		a.driftGuard.ClearHold()
		a.unit.SetDriftReading(10)
		a.observeDrift(10)
		if err := a.towerFSM.Apply(ctx, "drift_clear"); err != nil {
			return err
		}
	}
	a.persistSnapshot(model.TowerID(a.cfg.TowerID))
	return nil
}

func (a *App) observeDrift(ppm float64) {
	at := a.clk.Now()
	a.driftAgg.ObserveAll(ppm, at)
	a.unit.SetDriftReading(a.driftAgg.PeakPPM())
}

func (a *App) StatusLine() string {
	return a.unit.StatusLine(a.towerFSM.State(), len(a.fanCoord.Assignments()), len(a.sprayPlant.Assignments()))
}

func (a *App) TowerID() model.TowerID { return model.TowerID(a.cfg.TowerID) }

func (a *App) Snapshot() model.TowerSnapshot {
	snap, ok := a.mem.GetTower(model.TowerID(a.cfg.TowerID))
	if !ok {
		a.persistSnapshot(model.TowerID(a.cfg.TowerID))
		snap, _ = a.mem.GetTower(model.TowerID(a.cfg.TowerID))
	}
	return store.CloneSnapshot(snap)
}

func (a *App) Telemetry() model.TelemetryFrame {
	return model.TelemetryFrame{
		TowerID: model.TowerID(a.cfg.TowerID), DriftPPM: a.driftAgg.PeakPPM(),
		FanSpeed: a.fanCoord.AverageSpeed(), SprayGPM: a.sprayPlant.TotalFlow(),
		BasinTemp: a.unit.BasinTemperature(), State: a.towerFSM.State(), At: a.clk.Now(),
	}
}

func (a *App) AlarmManager() *alarms.Emitter { return a.alarms }

func (a *App) Config() config.Config { return a.cfg }

func (a *App) Clock() clock.Clock { return a.clk }
