package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

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
}

func Load() Config {
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
	}
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
