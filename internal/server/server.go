package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/agraja38/PiNetMonitor/internal/admin"
	"github.com/agraja38/PiNetMonitor/internal/config"
	"github.com/agraja38/PiNetMonitor/internal/reporting"
	"github.com/agraja38/PiNetMonitor/internal/store"
	"github.com/agraja38/PiNetMonitor/internal/system"
	"github.com/agraja38/PiNetMonitor/internal/version"
)

type Server struct {
	cfg    config.Config
	store  *store.Store
	mux    *http.ServeMux
	start  time.Time
	admin  *admin.Manager
}

func New(cfg config.Config, db *store.Store) *Server {
	s := &Server{
		cfg:   cfg,
		store: db,
		mux:   http.NewServeMux(),
		start: time.Now(),
		admin: admin.New(cfg),
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
	s.mux.HandleFunc("/api/v1/settings", s.handleSettings)
	s.mux.HandleFunc("/api/v1/update/status", s.handleUpdateStatus)
	s.mux.HandleFunc("/api/v1/update/start", s.handleUpdateStart)
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

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.admin.GetSettings()
		writeJSON(w, http.StatusOK, map[string]any{
			"http_addr":                cfg.HTTPAddr,
			"sample_interval":          cfg.SampleInterval.String(),
			"wan_interface":            cfg.WANInterface,
			"lan_interface":            cfg.LANInterface,
			"lan_cidr":                 cfg.LANCIDR,
			"gateway_mode":             cfg.GatewayMode,
			"enable_nat":               cfg.EnableNAT,
			"enable_https_attribution": cfg.EnableHTTPSAttribution,
			"github_repository":        cfg.GitHubRepository,
			"github_access_token_set":  cfg.GitHubAccessToken != "",
		})
	case http.MethodPut:
		var payload struct {
			HTTPAddr               string `json:"http_addr"`
			SampleInterval         string `json:"sample_interval"`
			WANInterface           string `json:"wan_interface"`
			LANInterface           string `json:"lan_interface"`
			LANCIDR                string `json:"lan_cidr"`
			GatewayMode            bool   `json:"gateway_mode"`
			EnableNAT              bool   `json:"enable_nat"`
			EnableHTTPSAttribution bool   `json:"enable_https_attribution"`
			GitHubRepository       string `json:"github_repository"`
			GitHubAccessToken      string `json:"github_access_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		cfg := s.admin.GetSettings()
		cfg.HTTPAddr = payload.HTTPAddr
		if payload.SampleInterval != "" {
			if duration, err := time.ParseDuration(payload.SampleInterval); err == nil {
				cfg.SampleInterval = duration
			}
		}
		cfg.WANInterface = payload.WANInterface
		cfg.LANInterface = payload.LANInterface
		cfg.LANCIDR = payload.LANCIDR
		cfg.GatewayMode = payload.GatewayMode
		cfg.EnableNAT = payload.EnableNAT
		cfg.EnableHTTPSAttribution = payload.EnableHTTPSAttribution
		if payload.GitHubRepository != "" {
			cfg.GitHubRepository = payload.GitHubRepository
		}
		cfg.GitHubAccessToken = payload.GitHubAccessToken
		if err := s.admin.SaveSettings(cfg); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	status := s.admin.RefreshUpdateStatus(r.Context())
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleUpdateStart(w http.ResponseWriter, _ *http.Request) {
	if err := s.admin.StartUpdate(); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	samples, err := s.store.LatestSamples()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	cfg := s.admin.GetSettings()

	payload := map[string]any{
		"name":       "PiNetMonitor",
		"tagline":    "PiNetMonitor – Simple internet usage monitoring for Raspberry Pi and Orange Pi.",
		"version":    version.Version,
		"commit":     version.Commit,
		"build_date": version.BuildDate,
		"uptime":     time.Since(s.start).Round(time.Second).String(),
		"gateway": map[string]any{
			"mode":                  cfg.GatewayMode,
			"nat_enabled":           cfg.EnableNAT,
			"https_attribution":     cfg.EnableHTTPSAttribution,
			"wan_interface":         cfg.WANInterface,
			"lan_interface":         cfg.LANInterface,
			"lan_cidr":              cfg.LANCIDR,
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
