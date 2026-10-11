// Package fetcher downloads property pages over plain HTTP.
package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// maxBodyBytes caps a page download at 5 MiB (5 << 20 = 5 * 1024 * 1024 =
// 5,242,880 bytes). Real property pages are ~80–105KB, so this leaves ~50x
// headroom while stopping a broken or hostile response from filling memory.
const maxBodyBytes = 5 << 20

// Fetcher returns the raw HTML of a page.
type Fetcher interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}

type HTTPFetcher struct {
	client    *http.Client
	userAgent string
}

func New(timeout time.Duration, userAgent string) *HTTPFetcher {
	return &HTTPFetcher{
		client:    &http.Client{Timeout: timeout},
		userAgent: userAgent,
	}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", f.userAgent)
	req.Header.Set("Accept", "text/html")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s: unexpected status %s", url, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", url, err)
	}
	if len(body) > maxBodyBytes {
		return nil, fmt.Errorf("get %s: page too large (over %d bytes)", url, maxBodyBytes)
	}
	return body, nil
}
