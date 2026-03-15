package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "status":
		run("systemctl", "status", "pinetmonitor.service", "--no-pager")
	case "restart":
		run("sudo", "systemctl", "restart", "pinetmonitor.service")
	case "logs":
		run("journalctl", "-u", "pinetmonitor.service", "-n", "100", "--no-pager")
	case "update":
		run("sudo", "/opt/pinetmonitor/scripts/update.sh")
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("PiNetMonitor CLI")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  pinetmonitor status")
	fmt.Println("  pinetmonitor restart")
	fmt.Println("  pinetmonitor logs")
	fmt.Println("  pinetmonitor update")
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "command failed: %v\n", err)
		os.Exit(1)
	}
}
