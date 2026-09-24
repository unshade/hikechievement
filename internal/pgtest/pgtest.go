// Package pgtest gives tests a real PostGIS database through testcontainers.
//
// Docker is the only requirement. A single container is started lazily on the
// first NewDatabase call and shared by the whole test binary; every test gets its
// own empty database inside it. The container is never stopped explicitly:
// testcontainers' reaper (Ryuk) removes it when the test process exits.
package pgtest

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	image        = "postgis/postgis:17-3.5"
	startTimeout = 3 * time.Minute
)

var (
	once      sync.Once
	container *postgres.PostgresContainer
	admin     *sql.DB
	config    *pgx.ConnConfig
	startErr  error
	dbCount   atomic.Int64
)

// NewDatabase returns a connection to a new, empty database (PostGIS available,
// no migrations applied). It is closed when the test ends.
func NewDatabase(t testing.TB) *sql.DB {
	t.Helper()

	once.Do(start)
	if startErr != nil {
		t.Fatalf("start postgres test container (is Docker running?): %v", startErr)
	}

	name := fmt.Sprintf("test_%d", dbCount.Add(1))
	if _, err := admin.ExecContext(t.Context(), "CREATE DATABASE "+name); err != nil {
		t.Fatal(err)
	}

	cfg := config.Copy()
	cfg.Database = name
	db := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func start() {
	// Not tied to a test's context: the container outlives the first test.
	ctx, cancel := context.WithTimeout(context.Background(), startTimeout)
	defer cancel()

	container, startErr = postgres.Run(ctx, image,
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("hike"),
		postgres.WithPassword("hike"),
		postgres.BasicWaitStrategies(),
	)
	if startErr != nil {
		return
	}
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		startErr = err
		return
	}
	if config, startErr = pgx.ParseConfig(dsn); startErr != nil {
		return
	}
	admin = stdlib.OpenDB(*config)
}
