package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/ranjabi/mamikos-watcher/internal/config"
	"github.com/ranjabi/mamikos-watcher/internal/fetcher"
	"github.com/ranjabi/mamikos-watcher/internal/watcher"
)

func main() {
	once := flag.Bool("once", false, "check every URL one time and exit, without the schedule")
	flag.Parse()

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("invalid config", "error", err.Error())
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	logger.Info("mamikos-watcher starting",
		"urls", len(cfg.URLs),
		"once", *once,
		"fetch_interval", cfg.FetchInterval.String(),
		"http_timeout", cfg.HTTPTimeout.String(),
	)

	w := watcher.New(fetcher.New(cfg.HTTPTimeout, cfg.UserAgent), cfg.URLs, logger)
	if *once {
		w.RunOnce(context.Background())
		return
	}

	// The scheduled loop arrives in step 7; until then, run a single check.
	w.RunOnce(context.Background())
}
