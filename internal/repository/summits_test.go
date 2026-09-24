package repository

import (
	"testing"

	"gorm.io/gorm"
)

func TestUpsertMany(t *testing.T) {
	get := func(t *testing.T, db *gorm.DB, id int64) Summit {
		t.Helper()
		s, err := gorm.G[Summit](db).Where("id = ?", id).First(t.Context())
		if err != nil {
			t.Fatalf("get summit %d: %v", id, err)
		}
		return s
	}

	t.Run("empty input does nothing", func(t *testing.T) {
		db := newTestDB(t)
		if err := NewSummitsRepository(db).UpsertMany(t.Context(), nil); err != nil {
			t.Fatal(err)
		}
		n, err := gorm.G[Summit](db).Count(t.Context(), "*")
		if err != nil || n != 0 {
			t.Errorf("count = %d, err = %v, want 0", n, err)
		}
	})

	t.Run("inserts a summit and reads it back", func(t *testing.T) {
		db := newTestDB(t)
		ele := float32(4478)
		err := NewSummitsRepository(db).UpsertMany(t.Context(), []Summit{{
			ID:         1,
			Names:      map[string]string{"default": "Matterhorn", "fr": "Cervin"},
			ElevationM: &ele,
			Geom:       Point{7.6586, 45.9763},
		}})
		if err != nil {
			t.Fatal(err)
		}

		got := get(t, db, 1)
		if got.Names["default"] != "Matterhorn" || got.Names["fr"] != "Cervin" {
			t.Errorf("names = %v", got.Names)
		}
		if got.ElevationM == nil || *got.ElevationM != 4478 {
			t.Errorf("elevation = %v, want 4478", got.ElevationM)
		}
		if got.Geom.Lon() != 7.6586 || got.Geom.Lat() != 45.9763 {
			t.Errorf("geom = lon %v lat %v, want lon 7.6586 lat 45.9763", got.Geom.Lon(), got.Geom.Lat())
		}
	})

	t.Run("stores unknown elevation as NULL", func(t *testing.T) {
		db := newTestDB(t)
		err := NewSummitsRepository(db).UpsertMany(t.Context(), []Summit{{
			ID:    2,
			Names: map[string]string{},
			Geom:  Point{1, 2},
		}})
		if err != nil {
			t.Fatal(err)
		}
		if got := get(t, db, 2); got.ElevationM != nil {
			t.Errorf("elevation = %v, want nil", *got.ElevationM)
		}
	})

	t.Run("updates an existing summit", func(t *testing.T) {
		db := newTestDB(t)
		repo := NewSummitsRepository(db)
		old := float32(100)
		if err := repo.UpsertMany(t.Context(), []Summit{{ID: 3, Names: map[string]string{"default": "Old"}, ElevationM: &old, Geom: Point{1, 2}}}); err != nil {
			t.Fatal(err)
		}

		updated := float32(200)
		if err := repo.UpsertMany(t.Context(), []Summit{{ID: 3, Names: map[string]string{"default": "New"}, ElevationM: &updated, Geom: Point{3, 4}}}); err != nil {
			t.Fatal(err)
		}

		got := get(t, db, 3)
		if got.Names["default"] != "New" || *got.ElevationM != 200 || got.Geom.Lon() != 3 || got.Geom.Lat() != 4 {
			t.Errorf("summit was not updated: %+v", got)
		}
		n, _ := gorm.G[Summit](db).Count(t.Context(), "*")
		if n != 1 {
			t.Errorf("count = %d, want 1", n)
		}
	})

	t.Run("inserts more summits than one batch", func(t *testing.T) {
		db := newTestDB(t)
		summits := make([]Summit, 2500)
		for i := range summits {
			summits[i] = Summit{ID: int64(i + 1), Names: map[string]string{}, Geom: Point{float64(i%180) - 90, float64(i%90) - 45}}
		}
		if err := NewSummitsRepository(db).UpsertMany(t.Context(), summits); err != nil {
			t.Fatal(err)
		}
		n, err := gorm.G[Summit](db).Count(t.Context(), "*")
		if err != nil || n != 2500 {
			t.Errorf("count = %d, err = %v, want 2500", n, err)
		}
	})
}
