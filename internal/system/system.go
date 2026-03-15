package system

import (
	"os"
	"runtime"
	"strings"
)

type RuntimeInfo struct {
	Hostname string `json:"hostname"`
	Kernel   string `json:"kernel"`
	Arch     string `json:"arch"`
	Platform string `json:"platform"`
}

func Runtime() RuntimeInfo {
	hostname, _ := os.Hostname()
	return RuntimeInfo{
		Hostname: hostname,
		Kernel:   runtime.GOOS,
		Arch:     runtime.GOARCH,
		Platform: detectPlatform(),
	}
}

func detectPlatform() string {
	data, err := os.ReadFile("/proc/device-tree/model")
	if err != nil {
		return "linux-sbc"
	}
	model := strings.TrimSpace(strings.ReplaceAll(string(data), "\x00", ""))
	if model == "" {
		return "linux-sbc"
	}
	return model
}
