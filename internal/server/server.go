package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/agraja38/PiNetMonitor/internal/config"
	"github.com/agraja38/PiNetMonitor/internal/reporting"
	"github.com/agraja38/PiNetMonitor/internal/store"
	"github.com/agraja38/PiNetMonitor/internal/system"
	"github.com/agraja38/PiNetMonitor/internal/version"
)

type Server struct {
	cfg   config.Config
	store *store.Store
	mux   *http.ServeMux
	start time.Time
}

func New(cfg config.Config, db *store.Store) *Server {
	s := &Server{
		cfg:   cfg,
		store: db,
		mux:   http.NewServeMux(),
		start: time.Now(),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/v1/status", s.handleStatus)
	s.mux.HandleFunc("/api/v1/reports/daily", s.handleDaily)
	s.mux.HandleFunc("/api/v1/reports/monthly", s.handleMonthly)
	s.mux.HandleFunc("/api/v1/attribution/daily-top", s.handleDailyTopServices)
	s.mux.HandleFunc("/api/v1/attribution/monthly-top", s.handleMonthlyTopServices)
	s.mux.HandleFunc("/api/v1/runtime", s.handleRuntime)

	frontendIndex := filepath.Join(s.cfg.FrontendDir, "index.html")
	if _, err := os.Stat(frontendIndex); err == nil {
		fileServer := http.FileServer(http.Dir(s.cfg.FrontendDir))
		s.mux.Handle("/", fileServer)
		return
	}
	s.mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "PiNetMonitor frontend not built. Run npm run build in web/.", http.StatusServiceUnavailable)
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	samples, err := s.store.LatestSamples()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	payload := map[string]any{
		"name":       "PiNetMonitor",
		"tagline":    "PiNetMonitor – Simple internet usage monitoring for Raspberry Pi and Orange Pi.",
		"version":    version.Version,
		"commit":     version.Commit,
		"build_date": version.BuildDate,
		"uptime":     time.Since(s.start).Round(time.Second).String(),
		"gateway": map[string]any{
			"mode":                  s.cfg.GatewayMode,
			"nat_enabled":           s.cfg.EnableNAT,
			"https_attribution":     s.cfg.EnableHTTPSAttribution,
			"wan_interface":         s.cfg.WANInterface,
			"lan_interface":         s.cfg.LANInterface,
			"lan_cidr":              s.cfg.LANCIDR,
			"attribution_strategy":  "dns+sni+flow-metadata-with-confidence",
			"capture_backend":       "proc-net-dev baseline with gateway extensions planned",
		},
		"samples": samples,
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleDaily(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.store.AggregateDailyTotals(14)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rows": rows})
}

func (s *Server) handleMonthly(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.store.AggregateMonthlyTotals(12)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rows": rows})
}

func (s *Server) handleDailyTopServices(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, reporting.EmptyServiceShareSummary("daily"))
}

func (s *Server) handleMonthlyTopServices(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, reporting.EmptyServiceShareSummary("monthly"))
}

func (s *Server) handleRuntime(w http.ResponseWriter, _ *http.Request) {
	summary, err := reporting.Build(s.store)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"runtime": system.Runtime(),
		"reports": summary,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
