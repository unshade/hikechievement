package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/unshade/hikechievement/internal/pgtest"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getHealth(t *testing.T, db *gorm.DB) (*http.Response, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	New(db).Handler().ServeHTTP(rec, req)

	res := rec.Result()
	var body map[string]any
	_ = json.NewDecoder(res.Body).Decode(&body)
	return res, body
}

func TestHealth(t *testing.T) {
	newDB := func(t *testing.T) *gorm.DB {
		t.Helper()
		sqlDB := pgtest.NewDatabase(t)
		db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
		if err != nil {
			t.Fatal(err)
		}
		return db
	}

	t.Run("ok when the database is reachable", func(t *testing.T) {
		res, body := getHealth(t, newDB(t))
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", res.StatusCode)
		}
		if body["status"] != "ok" {
			t.Errorf("body = %v, want status ok", body)
		}
	})

	t.Run("unavailable when the database is unreachable", func(t *testing.T) {
		db := newDB(t)
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		if err := sqlDB.Close(); err != nil {
			t.Fatal(err)
		}
		res, _ := getHealth(t, db)
		if res.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503", res.StatusCode)
		}
	})
}
