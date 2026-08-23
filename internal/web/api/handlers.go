package api

import (
	"encoding/json"
	"net/http"

	"github.com/lacsar712/cooltower/internal/app"
)

type Server struct{ app *app.App }

func NewServer(application *app.App) *Server { return &Server{app: application} }

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/snapshot", s.handleSnapshot)
	mux.HandleFunc("/telemetry", s.handleTelemetry)
	mux.HandleFunc("/alarms", s.handleAlarms)
	mux.HandleFunc("/tower/start", s.handleStart)
	mux.HandleFunc("/tower/stop", s.handleStop)
	mux.HandleFunc("/settings", s.handleSettings)
	return Chain(mux, RecoveryMiddleware, LoggingMiddleware, CORSMiddleware)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, http.StatusOK, HealthPayload{Status: "ok", Tower: s.app.TowerID().String()})
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, http.StatusOK, s.app.Snapshot())
}

func (s *Server) handleTelemetry(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, http.StatusOK, s.app.Telemetry())
}

func (s *Server) handleAlarms(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, http.StatusOK, s.app.AlarmManager().Active())
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if !MethodGuard(w, r, http.MethodPost) {
		return
	}
	if err := s.app.RunOnce(r.Context()); err != nil {
		ErrorResponse(w, http.StatusConflict, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"status": s.app.StatusLine()})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if !MethodGuard(w, r, http.MethodPost) {
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	cfg := s.app.Config()
	JSONResponse(w, http.StatusOK, SettingsPayload{
		TowerID: cfg.TowerID, FanCount: cfg.FanCount, SprayHeaders: cfg.SprayHeaderCount,
		DriftMaxPPM: cfg.DriftMaxPPM, DefaultSprayGPM: cfg.DefaultSprayGPM,
	})
}

type HealthPayload struct {
	Status string `json:"status"`
	Tower  string `json:"tower"`
}

type SettingsPayload struct {
	TowerID         string  `json:"tower_id"`
	FanCount        int     `json:"fan_count"`
	SprayHeaders    int     `json:"spray_headers"`
	DriftMaxPPM     float64 `json:"drift_max_ppm"`
	DefaultSprayGPM float64 `json:"default_spray_gpm"`
}

func MethodGuard(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	return false
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
