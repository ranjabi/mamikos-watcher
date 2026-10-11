// Package watcher checks each property once per run and logs the results.
package watcher

import (
	"context"
	"log/slog"
	"time"

	"github.com/ranjabi/mamikos-watcher/internal/fetcher"
	"github.com/ranjabi/mamikos-watcher/internal/parser"
)

const (
	statusError   = "error"
	statusUnknown = "Unknown"
)

// Report is one property's line in the run_complete summary.
type Report struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}

type Watcher struct {
	fetcher fetcher.Fetcher
	urls    []string
	logger  *slog.Logger
}

func New(f fetcher.Fetcher, urls []string, logger *slog.Logger) *Watcher {
	return &Watcher{fetcher: f, urls: urls, logger: logger}
}

// RunOnce checks every URL in order, logging a check line for each and a
// run_complete summary at the end. One failing URL doesn't stop the others.
func (w *Watcher) RunOnce(ctx context.Context) []Report {
	start := time.Now()
	reports := make([]Report, 0, len(w.urls))
	for _, url := range w.urls {
		reports = append(reports, w.check(ctx, url))
	}
	w.logger.Info("run_complete",
		"duration_ms", time.Since(start).Milliseconds(),
		"report", reports,
	)
	return reports
}

func (w *Watcher) check(ctx context.Context, url string) Report {
	start := time.Now()
	attrs := []any{"url", url}
	done := func(level slog.Level, title, status string, extra ...any) Report {
		if title == "" {
			title = url
		}
		attrs = append(attrs, "title", title)
		attrs = append(attrs, extra...)
		attrs = append(attrs, "status", status, "duration_ms", time.Since(start).Milliseconds())
		w.logger.Log(ctx, level, "check", attrs...)
		return Report{Title: title, Status: status}
	}

	page, err := w.fetcher.Fetch(ctx, url)
	if err != nil {
		return done(slog.LevelError, "", statusError, "error", err.Error())
	}
	d, err := parser.Extract(page)
	if err != nil {
		return done(slog.LevelWarn, d.Title, statusUnknown, "error", err.Error())
	}

	status := parser.Evaluate(d)
	level := slog.LevelInfo
	if status == statusUnknown {
		level = slog.LevelWarn
	}
	return done(level, d.Title, status, "available_room", d.AvailableRoom)
}
