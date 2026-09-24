package repository

import (
	"database/sql"
	"database/sql/driver"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/ewkb"
	"gorm.io/gorm/schema"
)

// srid is WGS 84 (plain latitude/longitude), the reference system of GPS and OSM.
const srid = 4326

// Point is a WGS 84 position stored as a PostGIS geometry(Point, 4326).
// Like orb.Point it is [longitude, latitude].
type Point orb.Point

var (
	_ driver.Valuer                = Point{}
	_ sql.Scanner                  = (*Point)(nil)
	_ schema.GormDataTypeInterface = Point{}
)

func (Point) GormDataType() string { return "geometry(Point,4326)" }

func (p Point) Lon() float64 { return p[0] }
func (p Point) Lat() float64 { return p[1] }

// Value writes the point as hex EWKB, which PostGIS reads as geometry text input.
func (p Point) Value() (driver.Value, error) {
	return ewkb.MarshalToHex(orb.Point(p), srid)
}

// Scan reads what PostgreSQL returns for a geometry column (hex EWKB).
func (p *Point) Scan(src any) error {
	if s, ok := src.(string); ok {
		src = []byte(s)
	}
	return ewkb.Scanner((*orb.Point)(p)).Scan(src)
}
