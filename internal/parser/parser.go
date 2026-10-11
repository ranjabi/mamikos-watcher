// Package parser reads the availability data Mamikos embeds in a property
// page and rebuilds the status text the page shows.
//
// The page renders the text with JavaScript, so it is not in the raw HTML.
// For now the text is derived from available_room only.
package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
)

// Detail is the part of the page's embedded `var detail = {...}` we need.
type Detail struct {
	Title         string
	AvailableRoom int
}

var (
	detailStart = regexp.MustCompile(`var\s+detail\s*=\s*`)
	titleTag    = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
)

// Extract finds the embedded detail object in a property page.
func Extract(page []byte) (Detail, error) {
	loc := detailStart.FindIndex(page)
	if loc == nil {
		return Detail{}, errors.New("detail data not found in page")
	}

	var raw struct {
		RoomTitle     string `json:"room_title"`
		AvailableRoom *int   `json:"available_room"`
	}
	// Decode reads one JSON value and ignores the JavaScript after it.
	if err := json.NewDecoder(bytes.NewReader(page[loc[1]:])).Decode(&raw); err != nil {
		return Detail{}, fmt.Errorf("decode detail data: %w", err)
	}
	if raw.AvailableRoom == nil {
		return Detail{}, errors.New("available_room missing from detail data")
	}

	d := Detail{
		Title:         strings.TrimSpace(raw.RoomTitle),
		AvailableRoom: *raw.AvailableRoom,
	}
	if d.Title == "" {
		if m := titleTag.FindSubmatch(page); m != nil {
			d.Title = strings.TrimSpace(html.UnescapeString(string(m[1])))
		}
	}
	return d, nil
}

// Evaluate returns the availability text the page shows for d.
func Evaluate(d Detail) string {
	switch n := d.AvailableRoom; {
	case n < 0:
		return "Unknown"
	case n == 0:
		return "Kamar penuh"
	case n > 3:
		return "Banyak pilihan kamar untukmu"
	default:
		return fmt.Sprintf("Tersisa %d kamar", n)
	}
}
