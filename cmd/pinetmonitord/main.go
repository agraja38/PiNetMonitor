package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agraja38/PiNetMonitor/internal/collector"
	"github.com/agraja38/PiNetMonitor/internal/config"
	"github.com/agraja38/PiNetMonitor/internal/server"
	"github.com/agraja38/PiNetMonitor/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	collectorService := collector.New(db, cfg.SampleInterval, []string{cfg.WANInterface, cfg.LANInterface})
	go func() {
		err := collectorService.Run(ctx)
		if err != nil && err != context.Canceled {
			log.Printf("collector stopped with error: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.New(cfg, db).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("PiNetMonitor listening on %s", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
