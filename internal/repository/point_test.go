package repository

import "testing"

// Point(1 2) with SRID 4326 as little-endian lowercase hex EWKB.
const point12Hex = "0101000020e6100000000000000000f03f0000000000000040"

func TestPointValue(t *testing.T) {
	t.Run("hex EWKB with SRID 4326", func(t *testing.T) {
		got, err := Point{1, 2}.Value()
		if err != nil {
			t.Fatal(err)
		}
		if got != point12Hex {
			t.Errorf("Value() = %v, want %v", got, point12Hex)
		}
	})
}

func TestPointScan(t *testing.T) {
	tests := []struct {
		name string
		src  any
	}{
		{"hex as string", point12Hex},
		{"hex as bytes", []byte(point12Hex)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Point
			if err := p.Scan(tt.src); err != nil {
				t.Fatal(err)
			}
			if p.Lon() != 1 || p.Lat() != 2 {
				t.Errorf("Scan() = lon %v lat %v, want lon 1 lat 2", p.Lon(), p.Lat())
			}
		})
	}

	t.Run("invalid data", func(t *testing.T) {
		var p Point
		if err := p.Scan("not hex"); err == nil {
			t.Error("expected an error")
		}
	})
}
