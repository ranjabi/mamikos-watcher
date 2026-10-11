package fetcher

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ranjabi/mamikos-watcher/internal/config"
	"github.com/ranjabi/mamikos-watcher/internal/parser"
)

// TestLiveSamplePages hits the real Mamikos site. Run with INTEGRATION=1.
// Room counts change over time, so it only checks the data is still there.
func TestLiveSamplePages(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("set INTEGRATION=1 to run against mamikos.com")
	}

	urls := []string{
		"https://mamikos.com/room/kost-kota-jakarta-selatan-kost-campur-eksklusif-kost-singgahsini-dstudio-mrt-haji-nawi-tipe-a-cilandak-jakarta-selatan-2",
		"https://mamikos.com/room/kost-kota-jakarta-selatan-kost-campur-eksklusif-kost-haji-nawi-residence-26a-tipe-vvip-cilandak-jakarta-selatan",
		"https://mamikos.com/room/kost-jakarta-selatan-kost-campur-eksklusif-kost-cipete-residence-tipe-superior-cilandak-jakarta-selatan-2",
	}
	f := New(30*time.Second, config.DefaultUserAgent)

	for _, url := range urls {
		t.Run(url[len("https://mamikos.com/room/"):], func(t *testing.T) {
			body, err := f.Fetch(context.Background(), url)
			if err != nil {
				t.Fatalf("fetch: %v", err)
			}
			d, err := parser.Extract(body)
			if err != nil {
				t.Fatalf("extract: %v", err)
			}
			if d.Title == "" {
				t.Error("empty title")
			}
			text := parser.Evaluate(d)
			if text == "Unknown" {
				t.Errorf("available_room = %d gave Unknown", d.AvailableRoom)
			}
			t.Logf("%s: available_room=%d -> %q", d.Title, d.AvailableRoom, text)
		})
	}
}
