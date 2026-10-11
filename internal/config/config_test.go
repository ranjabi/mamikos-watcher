package config

import (
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"
)

func envFrom(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"MAMIKOS_URLS": "https://mamikos.com/room/a",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Config{
		URLs:          []string{"https://mamikos.com/room/a"},
		FetchInterval: 30 * time.Minute,
		HTTPTimeout:   20 * time.Second,
		UserAgent:     DefaultUserAgent,
		LogLevel:      slog.LevelInfo,
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("got %+v\nwant %+v", cfg, want)
	}
}

func TestLoadAllSet(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"MAMIKOS_URLS":   " https://mamikos.com/room/a , ,https://mamikos.com/room/b?x=1 ",
		"FETCH_INTERVAL": "15m",
		"HTTP_TIMEOUT":   "5s",
		"USER_AGENT":     "my-agent",
		"LOG_LEVEL":      "debug",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Config{
		URLs:          []string{"https://mamikos.com/room/a", "https://mamikos.com/room/b?x=1"},
		FetchInterval: 15 * time.Minute,
		HTTPTimeout:   5 * time.Second,
		UserAgent:     "my-agent",
		LogLevel:      slog.LevelDebug,
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("got %+v\nwant %+v", cfg, want)
	}
}

func TestLoadErrors(t *testing.T) {
	const okURL = "https://mamikos.com/room/a"

	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"missing urls", map[string]string{}, "MAMIKOS_URLS is required"},
		{"only separators", map[string]string{"MAMIKOS_URLS": " , ,"}, "MAMIKOS_URLS is required"},
		{"not a url", map[string]string{"MAMIKOS_URLS": "not-a-url"}, `invalid URL "not-a-url"`},
		{"bad scheme", map[string]string{"MAMIKOS_URLS": "ftp://mamikos.com/x"}, `invalid URL "ftp://mamikos.com/x"`},
		{"bad interval", map[string]string{"MAMIKOS_URLS": okURL, "FETCH_INTERVAL": "soon"}, "FETCH_INTERVAL"},
		{"zero interval", map[string]string{"MAMIKOS_URLS": okURL, "FETCH_INTERVAL": "0s"}, "FETCH_INTERVAL must be positive"},
		{"negative timeout", map[string]string{"MAMIKOS_URLS": okURL, "HTTP_TIMEOUT": "-1s"}, "HTTP_TIMEOUT must be positive"},
		{"bad log level", map[string]string{"MAMIKOS_URLS": okURL, "LOG_LEVEL": "loud"}, "LOG_LEVEL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(envFrom(tt.env))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadReportsAllErrors(t *testing.T) {
	_, err := Load(envFrom(map[string]string{
		"FETCH_INTERVAL": "soon",
		"LOG_LEVEL":      "loud",
	}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	for _, want := range []string{"MAMIKOS_URLS", "FETCH_INTERVAL", "LOG_LEVEL"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %s", err, want)
		}
	}
}
