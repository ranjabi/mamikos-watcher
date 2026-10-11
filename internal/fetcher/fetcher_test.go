package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFetchReturnsBodyAndSendsUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Write([]byte("<html>ok</html>"))
	}))
	defer srv.Close()

	body, err := New(time.Second, "test-agent").Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != "<html>ok</html>" {
		t.Errorf("body = %q", body)
	}
	if gotUA != "test-agent" {
		t.Errorf("User-Agent = %q, want %q", gotUA, "test-agent")
	}
}

func TestFetchNon200IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := New(time.Second, "ua").Fetch(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("error = %v, want it to mention 403", err)
	}
}

func TestFetchTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(time.Second):
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()

	start := time.Now()
	_, err := New(50*time.Millisecond, "ua").Fetch(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("took %s, timeout not applied", elapsed)
	}
}

func TestFetchCancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New(time.Second, "ua").Fetch(ctx, srv.URL); err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestFetchBodyTooLarge(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("x", maxBodyBytes+1)))
	}))
	defer srv.Close()

	_, err := New(time.Second, "ua").Fetch(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Errorf("error = %v, want it to mention too large", err)
	}
}
