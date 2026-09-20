package importer

import (
	"testing"

	"github.com/paulmach/osm"
)

func TestParseElevation(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want *float64
	}{
		{"integer", "4808", new(4808.0)},
		{"decimal point", "4808.5", new(4808.5)},
		{"decimal comma", "4808,5", new(4808.5)},
		{"meter suffix", "4808 m", new(4808.0)},
		{"thousands space", "4 808", new(4808.0)},
		{"surrounding spaces", " 2002 ", new(2002.0)},
		{"negative", "-100", new(-100.0)},
		{"empty", "", nil},
		{"not a number", "abc", nil},
		{"feet", "15000 ft", nil},
		{"above range", "48080", nil},
		{"NaN", "NaN", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseElevation(tt.in)
			switch {
			case got == nil && tt.want == nil:
			case got == nil || tt.want == nil || *got != *tt.want:
				t.Errorf("parseElevation(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNames(t *testing.T) {
	tests := []struct {
		name string
		tags osm.Tags
		want map[string]string
	}{
		{
			"default name only",
			osm.Tags{{Key: "name", Value: "Matterhorn"}},
			map[string]string{DefaultName: "Matterhorn"},
		},
		{
			"several languages",
			osm.Tags{
				{Key: "name", Value: "Matterhorn"},
				{Key: "name:fr", Value: "Cervin"},
				{Key: "name:it", Value: "Cervino"},
				{Key: "name:zh-Hans", Value: "马特洪峰"},
			},
			map[string]string{DefaultName: "Matterhorn", "fr": "Cervin", "it": "Cervino", "zh-Hans": "马特洪峰"},
		},
		{
			"non language name tags ignored",
			osm.Tags{
				{Key: "name:etymology", Value: "x"},
				{Key: "name:left", Value: "x"},
				{Key: "alt_name", Value: "x"},
			},
			map[string]string{},
		},
		{
			"empty value ignored",
			osm.Tags{{Key: "name:en", Value: ""}},
			map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := names(tt.tags)
			if len(got) != len(tt.want) {
				t.Fatalf("names() = %v, want %v", got, tt.want)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("names()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
