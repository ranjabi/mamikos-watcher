package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		rooms int
		want  string
	}{
		{-1, "Unknown"},
		{0, "Kamar penuh"},
		{1, "Tersisa 1 kamar"},
		{2, "Tersisa 2 kamar"},
		{3, "Tersisa 3 kamar"},
		{4, "Banyak pilihan kamar untukmu"},
		{20, "Banyak pilihan kamar untukmu"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := Evaluate(Detail{AvailableRoom: tt.rooms}); got != tt.want {
				t.Errorf("Evaluate(%d rooms) = %q, want %q", tt.rooms, got, tt.want)
			}
		})
	}
}

func TestExtractSamplePages(t *testing.T) {
	tests := []struct {
		file string
		want Detail
		text string
	}{
		{"kamar_penuh.html", Detail{
			Title:         "Kost Singgahsini Dstudio MRT Haji Nawi Tipe A Cilandak Jakarta Selatan",
			AvailableRoom: 0,
		}, "Kamar penuh"},
		{"tersisa_2_kamar.html", Detail{
			Title:         "Kost Haji Nawi Residence 26A Tipe VVIP Cilandak Jakarta Selatan",
			AvailableRoom: 2,
		}, "Tersisa 2 kamar"},
		{"banyak_pilihan.html", Detail{
			Title:         "Kost Cipete Residence Tipe Superior Cilandak Jakarta Selatan",
			AvailableRoom: 4,
		}, "Banyak pilihan kamar untukmu"},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			html, err := os.ReadFile(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			got, err := Extract(html)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v\nwant %+v", got, tt.want)
			}
			if text := Evaluate(got); text != tt.text {
				t.Errorf("Evaluate text = %q, want %q", text, tt.text)
			}
		})
	}
}

func TestExtractTitleFallsBackToTitleTag(t *testing.T) {
	html := `<title>Kost &amp; Co Tebet</title><script>var detail = {"room_title":"","available_room":1};</script>`
	got, err := Extract([]byte(html))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "Kost & Co Tebet" {
		t.Errorf("Title = %q, want %q", got.Title, "Kost & Co Tebet")
	}
}

func TestExtractErrors(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantErr string
	}{
		{"no detail object", `<title>x</title><div id="app"></div>`, "detail data not found"},
		{"broken json", `<script>var detail = {"available_room":</script>`, "decode detail data"},
		{"no available_room", `<script>var detail = {"room_title":"x"};</script>`, "available_room missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Extract([]byte(tt.html))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
