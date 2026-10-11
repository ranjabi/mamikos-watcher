// Package config loads the watcher's settings from environment variables.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"
)

// DefaultUserAgent is a desktop Chrome string; some sites block Go's default User-Agent.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"

const (
	defaultFetchInterval = 30 * time.Minute
	defaultHTTPTimeout   = 20 * time.Second
)

type Config struct {
	URLs          []string
	FetchInterval time.Duration
	HTTPTimeout   time.Duration
	UserAgent     string
	LogLevel      slog.Level
}

// Load reads the config using getenv (normally os.Getenv) and reports
// every invalid setting at once.
func Load(getenv func(string) string) (Config, error) {
	var errs []error

	urls, err := parseURLs(getenv("MAMIKOS_URLS"))
	errs = append(errs, err)

	interval, err := parseDuration("FETCH_INTERVAL", getenv("FETCH_INTERVAL"), defaultFetchInterval)
	errs = append(errs, err)

	timeout, err := parseDuration("HTTP_TIMEOUT", getenv("HTTP_TIMEOUT"), defaultHTTPTimeout)
	errs = append(errs, err)

	userAgent := strings.TrimSpace(getenv("USER_AGENT"))
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}

	var level slog.Level
	if v := strings.TrimSpace(getenv("LOG_LEVEL")); v != "" {
		if err := level.UnmarshalText([]byte(v)); err != nil {
			errs = append(errs, fmt.Errorf("LOG_LEVEL: %q is not one of debug, info, warn, error", v))
		}
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return Config{
		URLs:          urls,
		FetchInterval: interval,
		HTTPTimeout:   timeout,
		UserAgent:     userAgent,
		LogLevel:      level,
	}, nil
}

func parseURLs(raw string) ([]string, error) {
	var urls []string
	var errs []error
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		u, err := url.Parse(part)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			errs = append(errs, fmt.Errorf("MAMIKOS_URLS: invalid URL %q (must be http or https)", part))
			continue
		}
		urls = append(urls, part)
	}
	if len(urls) == 0 && len(errs) == 0 {
		return nil, errors.New("MAMIKOS_URLS is required")
	}
	return urls, errors.Join(errs...)
}

func parseDuration(name, raw string, def time.Duration) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is not a duration (e.g. 30s, 15m, 1h)", name, raw)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s must be positive, got %s", name, raw)
	}
	return d, nil
}
