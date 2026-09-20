// Package importer reads summits out of OpenStreetMap extracts.
package importer

import (
	"context"
	"fmt"
	"io"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

// DefaultName is the key of the plain `name` tag in Peak.Names.
const DefaultName = "default"

// Peak is a natural=peak node.
type Peak struct {
	ID         int64
	Names      map[string]string // language code (or DefaultName) -> name
	ElevationM *float64          // nil when the ele tag is missing or unusable
	Lat, Lon   float64
}

// ReadPeaks streams a .osm.pbf and calls fn for every natural=peak node.
// Only nodes are read: peaks mapped as ways or relations are ignored.
func ReadPeaks(ctx context.Context, r io.Reader, fn func(Peak) error) error {
	scan := osmpbf.New(ctx, r, runtime.GOMAXPROCS(-1))
	defer scan.Close()

	scan.SkipWays = true
	scan.SkipRelations = true
	scan.FilterNode = func(n *osm.Node) bool { return n.Tags.Find("natural") == "peak" }

	for scan.Scan() {
		n, ok := scan.Object().(*osm.Node)
		if !ok {
			continue
		}
		if err := fn(peakFromNode(n)); err != nil {
			return err
		}
	}
	if err := scan.Err(); err != nil {
		return fmt.Errorf("read pbf: %w", err)
	}
	return nil
}

func peakFromNode(n *osm.Node) Peak {
	return Peak{
		ID:         int64(n.ID),
		Names:      names(n.Tags),
		ElevationM: parseElevation(n.Tags.Find("ele")),
		Lat:        n.Lat,
		Lon:        n.Lon,
	}
}

// langTag matches name:<lang> keys such as name:fr or name:zh-Hans, but not
// name:left, name:etymology, etc.
var langTag = regexp.MustCompile(`^name:([a-z]{2,3}(?:-[A-Za-z0-9]+)*)$`)

func names(tags osm.Tags) map[string]string {
	out := make(map[string]string)
	for _, t := range tags {
		if t.Value == "" {
			continue
		}
		if t.Key == "name" {
			out[DefaultName] = t.Value
		} else if m := langTag.FindStringSubmatch(t.Key); m != nil {
			out[m[1]] = t.Value
		}
	}
	return out
}

// Plausible range for a summit elevation in meters (Dead Sea shore to above Everest).
const minEle, maxEle = -500, 9000

// parseElevation reads an OSM ele tag ("4808", "4808.5", "4 808 m", "4808,5").
// It returns nil for anything it can't confidently read as meters.
func parseElevation(s string) *float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "m")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.Replace(s, ",", ".", 1)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < minEle || v > maxEle {
		return nil
	}
	return &v
}
