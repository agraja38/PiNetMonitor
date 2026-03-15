package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agraja38/PiNetMonitor/internal/config"
	"github.com/agraja38/PiNetMonitor/internal/version"
)

const (
	updateLogPath = "/var/log/pinetmonitor/update-web.log"
	updateScript  = "/opt/pinetmonitor/scripts/update.sh"
	configScript  = "/opt/pinetmonitor/scripts/configure-network.sh"
	updateUnit    = "pinetmonitor-web-update.service"
)

type UpdateStatus struct {
	Running         bool      `json:"running"`
	LastStartedAt   time.Time `json:"last_started_at,omitempty"`
	LastFinishedAt  time.Time `json:"last_finished_at,omitempty"`
	LastExitCode    int       `json:"last_exit_code"`
	Log             string    `json:"log"`
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	CheckError      string    `json:"check_error,omitempty"`
}

type Manager struct {
	mu         sync.Mutex
	cfg        config.Config
	update     UpdateStatus
	httpClient *http.Client
}

func New(cfg config.Config) *Manager {
	return &Manager{
		cfg: cfg,
		update: UpdateStatus{
			CurrentVersion: version.Version,
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *Manager) GetSettings() config.Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg
}

func (m *Manager) SaveSettings(next config.Config) error {
	if err := next.Validate(); err != nil {
		return err
	}
	if err := config.WriteEnvFile(config.EnvFilePath, next); err != nil {
		return err
	}

	cmd := exec.Command("/bin/sh", "-lc", fmt.Sprintf("set -a && . %s && set +a && %s", config.EnvFilePath, configScript))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apply network configuration: %v: %s", err, strings.TrimSpace(string(output)))
	}

	m.mu.Lock()
	m.cfg = next
	m.mu.Unlock()
	return nil
}

func (m *Manager) RefreshUpdateStatus(ctx context.Context) UpdateStatus {
	m.mu.Lock()
	status := m.update
	cfg := m.cfg
	m.mu.Unlock()

	latest, err := m.fetchLatestVersion(ctx, cfg)
	status.CurrentVersion = version.Version
	status.LatestVersion = latest
	status.UpdateAvailable = compareVersions(latest, version.Version) > 0
	if err != nil {
		status.CheckError = err.Error()
	}

	if logBytes, readErr := os.ReadFile(updateLogPath); readErr == nil {
		status.Log = string(logBytes)
	}

	activeState, exitCode := updateUnitState(ctx)
	status.Running = activeState == "active" || activeState == "activating"
	if !status.Running && exitCode >= 0 {
		status.LastExitCode = exitCode
	}

	m.mu.Lock()
	m.update = status
	defer m.mu.Unlock()
	return m.update
}

func (m *Manager) StartUpdate() error {
	m.mu.Lock()
	if m.update.Running {
		m.mu.Unlock()
		return fmt.Errorf("update already running")
	}
	m.update.Running = true
	m.update.LastStartedAt = time.Now()
	m.update.Log = ""
	m.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(updateLogPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(updateLogPath, []byte(""), 0o644); err != nil {
		return err
	}

	command := fmt.Sprintf("export GITHUB_ACCESS_TOKEN=%q; %s >> %s 2>&1", m.cfg.GitHubAccessToken, updateScript, updateLogPath)
	cmd := exec.Command("systemd-run", "--unit", strings.TrimSuffix(updateUnit, ".service"), "--collect", "/bin/sh", "-lc", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		m.finishUpdate(1)
		m.appendLog(string(output))
		return err
	}
	m.appendLog(string(output))

	return nil
}

func (m *Manager) finishUpdate(exitCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.update.Running = false
	m.update.LastFinishedAt = time.Now()
	m.update.LastExitCode = exitCode
	if logBytes, err := os.ReadFile(updateLogPath); err == nil {
		m.update.Log = string(logBytes)
	}
}

func (m *Manager) appendLog(text string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return
	}
	current, _ := os.ReadFile(updateLogPath)
	combined := string(current)
	if combined != "" && !strings.HasSuffix(combined, "\n") {
		combined += "\n"
	}
	combined += trimmed + "\n"
	_ = os.WriteFile(updateLogPath, []byte(combined), 0o644)
}

func (m *Manager) fetchLatestVersion(ctx context.Context, cfg config.Config) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", cfg.GitHubRepository)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := strings.TrimSpace(cfg.GitHubAccessToken); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var body bytes.Buffer
		_, _ = body.ReadFrom(resp.Body)
		return "", fmt.Errorf("github release check failed: %s", strings.TrimSpace(body.String()))
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(payload.TagName), "v"), nil
}

func compareVersions(a, b string) int {
	if a == "" && b == "" {
		return 0
	}
	parse := func(input string) []int {
		parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(input), "v"), ".")
		result := make([]int, 0, len(parts))
		for _, part := range parts {
			n, _ := strconv.Atoi(part)
			result = append(result, n)
		}
		return result
	}
	left := parse(a)
	right := parse(b)
	maxLen := len(left)
	if len(right) > maxLen {
		maxLen = len(right)
	}
	for i := 0; i < maxLen; i++ {
		lv, rv := 0, 0
		if i < len(left) {
			lv = left[i]
		}
		if i < len(right) {
			rv = right[i]
		}
		if lv > rv {
			return 1
		}
		if lv < rv {
			return -1
		}
	}
	return 0
}

func updateUnitState(ctx context.Context) (string, int) {
	cmd := exec.CommandContext(ctx, "systemctl", "show", updateUnit, "--property=ActiveState", "--property=ExecMainStatus", "--value")
	output, err := cmd.Output()
	if err != nil {
		return "inactive", -1
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	state := "inactive"
	exitCode := -1
	if len(lines) >= 1 && strings.TrimSpace(lines[0]) != "" {
		state = strings.TrimSpace(lines[0])
	}
	if len(lines) >= 2 {
		if parsed, parseErr := strconv.Atoi(strings.TrimSpace(lines[1])); parseErr == nil {
			exitCode = parsed
		}
	}
	return state, exitCode
}
