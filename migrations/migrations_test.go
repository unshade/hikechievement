package migrations

import (
	"strings"
	"testing"

	"github.com/unshade/hikechievement/internal/pgtest"
)

func TestUp(t *testing.T) {
	t.Run("creates the postgis extension", func(t *testing.T) {
		db := pgtest.NewDatabase(t)
		if err := Up(t.Context(), db); err != nil {
			t.Fatal(err)
		}

		var n int
		err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM pg_extension WHERE extname = 'postgis'").Scan(&n)
		if err != nil || n != 1 {
			t.Errorf("postgis extension count = %d, err = %v, want 1", n, err)
		}
	})

	t.Run("creates the summits table", func(t *testing.T) {
		db := pgtest.NewDatabase(t)
		if err := Up(t.Context(), db); err != nil {
			t.Fatal(err)
		}

		rows, err := db.QueryContext(t.Context(),
			"SELECT column_name, udt_name, is_nullable FROM information_schema.columns WHERE table_name = 'summits'")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		type column struct{ udt, nullable string }
		got := map[string]column{}
		for rows.Next() {
			var name string
			var c column
			if err := rows.Scan(&name, &c.udt, &c.nullable); err != nil {
				t.Fatal(err)
			}
			got[name] = c
		}

		want := map[string]column{
			"id":          {"int8", "NO"},
			"names":       {"jsonb", "NO"},
			"elevation_m": {"float4", "YES"},
			"geom":        {"geometry", "NO"},
		}
		if len(got) != len(want) {
			t.Fatalf("columns = %v, want %v", got, want)
		}
		for name, w := range want {
			if got[name] != w {
				t.Errorf("column %s = %+v, want %+v", name, got[name], w)
			}
		}
	})

	t.Run("creates a gist index on geom", func(t *testing.T) {
		db := pgtest.NewDatabase(t)
		if err := Up(t.Context(), db); err != nil {
			t.Fatal(err)
		}

		var def string
		err := db.QueryRowContext(t.Context(),
			"SELECT indexdef FROM pg_indexes WHERE tablename = 'summits' AND indexname = 'summits_geom_idx'").Scan(&def)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(def, "USING gist (geom)") {
			t.Errorf("indexdef = %q, want a gist index on geom", def)
		}
	})

	t.Run("can run twice", func(t *testing.T) {
		db := pgtest.NewDatabase(t)
		for range 2 {
			if err := Up(t.Context(), db); err != nil {
				t.Fatal(err)
			}
		}
	})
}
