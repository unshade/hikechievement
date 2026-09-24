// Package migrations holds the goose SQL migrations, embedded in the binary.
package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var files embed.FS

// Up applies all pending migrations.
func Up(ctx context.Context, db *sql.DB) error {
	provider, err := goose.NewProvider(goose.DialectPostgres, db, files)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
