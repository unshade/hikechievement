package repository

import (
	"testing"

	"github.com/unshade/hikechievement/internal/pgtest"
	"github.com/unshade/hikechievement/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newTestDB returns a GORM connection to a new, migrated database.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB := pgtest.NewDatabase(t)
	if err := migrations.Up(t.Context(), sqlDB); err != nil {
		t.Fatal(err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db
}
