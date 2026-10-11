package main

import (
	"log/slog"
	"os"

	"github.com/ranjabi/mamikos-watcher/internal/config"
)

func main() {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("invalid config", "error", err.Error())
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	logger.Info("mamikos-watcher starting",
		"urls", len(cfg.URLs),
		"fetch_interval", cfg.FetchInterval.String(),
		"http_timeout", cfg.HTTPTimeout.String(),
	)
}
