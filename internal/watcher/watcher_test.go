package watcher

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"strconv"
	"testing"
)

// fakeFetcher returns canned pages or errors per URL.
type fakeFetcher map[string]any

func (f fakeFetcher) Fetch(_ context.Context, url string) ([]byte, error) {
	switch v := f[url].(type) {
	case string:
		return []byte(v), nil
	case error:
		return nil, v
	}
	return nil, errors.New("no fake for " + url)
}

func page(title string, rooms int) string {
	return `<script>var detail = {"room_title":"` + title + `","available_room":` + strconv.Itoa(rooms) + `};</script>`
}

func run(f fakeFetcher, urls ...string) []Report {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(f, urls, logger).RunOnce(context.Background())
}

func TestRunOnce(t *testing.T) {
	tests := []struct {
		name  string
		pages fakeFetcher
		urls  []string
		want  []Report
	}{
		{
			name: "each url in order",
			pages: fakeFetcher{
				"https://m/a": page("Kost A", 0),
				"https://m/b": page("Kost B", 2),
				"https://m/c": page("Kost C", 7),
			},
			urls: []string{"https://m/a", "https://m/b", "https://m/c"},
			want: []Report{
				{Title: "Kost A", Status: "Kamar penuh"},
				{Title: "Kost B", Status: "Tersisa 2 kamar"},
				{Title: "Kost C", Status: "Banyak pilihan kamar untukmu"},
			},
		},
		{
			name: "fetch error doesn't stop the rest",
			pages: fakeFetcher{
				"https://m/down": errors.New("connection refused"),
				"https://m/b":    page("Kost B", 1),
			},
			urls: []string{"https://m/down", "https://m/b"},
			want: []Report{
				{Title: "https://m/down", Status: "error"},
				{Title: "Kost B", Status: "Tersisa 1 kamar"},
			},
		},
		{
			name:  "page without data is unknown",
			pages: fakeFetcher{"https://m/x": "<html>no data here</html>"},
			urls:  []string{"https://m/x"},
			want:  []Report{{Title: "https://m/x", Status: "Unknown"}},
		},
		{
			name:  "negative rooms is unknown",
			pages: fakeFetcher{"https://m/n": page("Kost N", -1)},
			urls:  []string{"https://m/n"},
			want:  []Report{{Title: "Kost N", Status: "Unknown"}},
		},
		{
			name:  "empty title falls back to url",
			pages: fakeFetcher{"https://m/t": page("", 3)},
			urls:  []string{"https://m/t"},
			want:  []Report{{Title: "https://m/t", Status: "Tersisa 3 kamar"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := run(tt.pages, tt.urls...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v\nwant %+v", got, tt.want)
			}
		})
	}
}
