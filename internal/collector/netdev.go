package collector

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/agraja38/PiNetMonitor/internal/store"
)

type Collector struct {
	store         *store.Store
	sampleEvery   time.Duration
	watchIfaces   map[string]struct{}
}

func New(db *store.Store, sampleEvery time.Duration, ifaces []string) *Collector {
	watch := make(map[string]struct{}, len(ifaces))
	for _, iface := range ifaces {
		if iface == "" {
			continue
		}
		watch[iface] = struct{}{}
	}
	return &Collector{
		store:       db,
		sampleEvery: sampleEvery,
		watchIfaces: watch,
	}
}

func (c *Collector) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.sampleEvery)
	defer ticker.Stop()

	if err := c.sample(); err != nil {
		log.Printf("initial sample failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := c.sample(); err != nil {
				log.Printf("sample failed: %v", err)
			}
		}
	}
}

func (c *Collector) sample() error {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return fmt.Errorf("open /proc/net/dev: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	now := time.Now().UTC()
	for scanner.Scan() {
		lineNo++
		if lineNo <= 2 {
			continue
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if _, ok := c.watchIfaces[iface]; !ok {
			continue
		}

		fields := strings.Fields(strings.TrimSpace(parts[1]))
		if len(fields) < 16 {
			continue
		}
		rxBytes, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return err
		}
		txBytes, err := strconv.ParseInt(fields[8], 10, 64)
		if err != nil {
			return err
		}

		if err := c.store.InsertSample(store.Sample{
			Timestamp: now,
			Interface: iface,
			RxBytes:   rxBytes,
			TxBytes:   txBytes,
		}); err != nil {
			return err
		}
	}
	return scanner.Err()
}
