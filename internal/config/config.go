package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const EnvFilePath = "/etc/pinetmonitor/pinetmonitor.env"

type Config struct {
	HTTPAddr               string
	DBPath                 string
	FrontendDir            string
	SampleInterval         time.Duration
	WANInterface           string
	LANInterface           string
	LANCIDR                string
	GatewayMode            bool
	EnableNAT              bool
	EnableHTTPSAttribution bool
	GitHubUsername         string
	GitHubRepository       string
	GitHubAccessToken      string
}

func Load() Config {
	cfg := Default()
	if fileValues, err := ReadEnvFile(EnvFilePath); err == nil {
		applyValues(&cfg, fileValues)
	}
	applyValues(&cfg, map[string]string{
		"PINETMONITOR_HTTP_ADDR":               os.Getenv("PINETMONITOR_HTTP_ADDR"),
		"PINETMONITOR_DB_PATH":                 os.Getenv("PINETMONITOR_DB_PATH"),
		"PINETMONITOR_FRONTEND_DIR":            os.Getenv("PINETMONITOR_FRONTEND_DIR"),
		"PINETMONITOR_SAMPLE_INTERVAL":         os.Getenv("PINETMONITOR_SAMPLE_INTERVAL"),
		"PINETMONITOR_WAN_IFACE":               os.Getenv("PINETMONITOR_WAN_IFACE"),
		"PINETMONITOR_LAN_IFACE":               os.Getenv("PINETMONITOR_LAN_IFACE"),
		"PINETMONITOR_LAN_CIDR":                os.Getenv("PINETMONITOR_LAN_CIDR"),
		"PINETMONITOR_GATEWAY_MODE":            os.Getenv("PINETMONITOR_GATEWAY_MODE"),
		"PINETMONITOR_ENABLE_NAT":              os.Getenv("PINETMONITOR_ENABLE_NAT"),
		"PINETMONITOR_ENABLE_HTTPS_ATTRIBUTION": os.Getenv("PINETMONITOR_ENABLE_HTTPS_ATTRIBUTION"),
		"PINETMONITOR_GITHUB_USERNAME":         os.Getenv("PINETMONITOR_GITHUB_USERNAME"),
		"PINETMONITOR_GITHUB_REPOSITORY":       os.Getenv("PINETMONITOR_GITHUB_REPOSITORY"),
		"GITHUB_ACCESS_TOKEN":                  os.Getenv("GITHUB_ACCESS_TOKEN"),
	})
	return cfg
}

func Default() Config {
	return Config{
		HTTPAddr:               getEnv("PINETMONITOR_HTTP_ADDR", "0.0.0.0:8080"),
		DBPath:                 getEnv("PINETMONITOR_DB_PATH", "/var/lib/pinetmonitor/pinetmonitor.db"),
		FrontendDir:            getEnv("PINETMONITOR_FRONTEND_DIR", "/opt/pinetmonitor/web/dist"),
		SampleInterval:         getDuration("PINETMONITOR_SAMPLE_INTERVAL", 30*time.Second),
		WANInterface:           getEnv("PINETMONITOR_WAN_IFACE", "eth0"),
		LANInterface:           getEnv("PINETMONITOR_LAN_IFACE", "eth1"),
		LANCIDR:                getEnv("PINETMONITOR_LAN_CIDR", "192.168.50.0/24"),
		GatewayMode:            getBool("PINETMONITOR_GATEWAY_MODE", true),
		EnableNAT:              getBool("PINETMONITOR_ENABLE_NAT", true),
		EnableHTTPSAttribution: getBool("PINETMONITOR_ENABLE_HTTPS_ATTRIBUTION", true),
		GitHubUsername:         getEnv("PINETMONITOR_GITHUB_USERNAME", "agraja38"),
		GitHubRepository:       getEnv("PINETMONITOR_GITHUB_REPOSITORY", "agraja38/PiNetMonitor"),
		GitHubAccessToken:      strings.TrimSpace(os.Getenv("GITHUB_ACCESS_TOKEN")),
	}
}

func ReadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		values[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return values, scanner.Err()
}

func WriteEnvFile(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content := strings.Join([]string{
		"PINETMONITOR_HTTP_ADDR=" + cfg.HTTPAddr,
		"PINETMONITOR_DB_PATH=" + cfg.DBPath,
		"PINETMONITOR_FRONTEND_DIR=" + cfg.FrontendDir,
		"PINETMONITOR_SAMPLE_INTERVAL=" + cfg.SampleInterval.String(),
		"PINETMONITOR_WAN_IFACE=" + cfg.WANInterface,
		"PINETMONITOR_LAN_IFACE=" + cfg.LANInterface,
		"PINETMONITOR_LAN_CIDR=" + cfg.LANCIDR,
		"PINETMONITOR_GATEWAY_MODE=" + strconv.FormatBool(cfg.GatewayMode),
		"PINETMONITOR_ENABLE_NAT=" + strconv.FormatBool(cfg.EnableNAT),
		"PINETMONITOR_ENABLE_HTTPS_ATTRIBUTION=" + strconv.FormatBool(cfg.EnableHTTPSAttribution),
		"PINETMONITOR_GITHUB_USERNAME=" + cfg.GitHubUsername,
		"PINETMONITOR_GITHUB_REPOSITORY=" + cfg.GitHubRepository,
		"GITHUB_ACCESS_TOKEN=" + cfg.GitHubAccessToken,
		"",
	}, "\n")
	return os.WriteFile(path, []byte(content), 0o600)
}

func applyValues(cfg *Config, values map[string]string) {
	if value := strings.TrimSpace(values["PINETMONITOR_HTTP_ADDR"]); value != "" {
		cfg.HTTPAddr = value
	}
	if value := strings.TrimSpace(values["PINETMONITOR_DB_PATH"]); value != "" {
		cfg.DBPath = value
	}
	if value := strings.TrimSpace(values["PINETMONITOR_FRONTEND_DIR"]); value != "" {
		cfg.FrontendDir = value
	}
	if value := strings.TrimSpace(values["PINETMONITOR_SAMPLE_INTERVAL"]); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			cfg.SampleInterval = parsed
		}
	}
	if value := strings.TrimSpace(values["PINETMONITOR_WAN_IFACE"]); value != "" {
		cfg.WANInterface = value
	}
	if value := strings.TrimSpace(values["PINETMONITOR_LAN_IFACE"]); value != "" {
		cfg.LANInterface = value
	}
	if value := strings.TrimSpace(values["PINETMONITOR_LAN_CIDR"]); value != "" {
		cfg.LANCIDR = value
	}
	if value := strings.TrimSpace(values["PINETMONITOR_GATEWAY_MODE"]); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			cfg.GatewayMode = parsed
		}
	}
	if value := strings.TrimSpace(values["PINETMONITOR_ENABLE_NAT"]); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			cfg.EnableNAT = parsed
		}
	}
	if value := strings.TrimSpace(values["PINETMONITOR_ENABLE_HTTPS_ATTRIBUTION"]); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			cfg.EnableHTTPSAttribution = parsed
		}
	}
	if value := strings.TrimSpace(values["PINETMONITOR_GITHUB_USERNAME"]); value != "" {
		cfg.GitHubUsername = value
	}
	if value := strings.TrimSpace(values["PINETMONITOR_GITHUB_REPOSITORY"]); value != "" {
		cfg.GitHubRepository = value
	}
	if value := strings.TrimSpace(values["GITHUB_ACCESS_TOKEN"]); value != "" {
		cfg.GitHubAccessToken = value
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.WANInterface) == "" {
		return fmt.Errorf("WAN interface is required")
	}
	if strings.TrimSpace(c.LANInterface) == "" {
		return fmt.Errorf("LAN interface is required")
	}
	if strings.TrimSpace(c.LANCIDR) == "" {
		return fmt.Errorf("LAN CIDR is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
