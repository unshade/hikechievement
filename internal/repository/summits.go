// Package repository is the data access layer. Consumers depend on the
// interfaces declared here, never on the GORM implementations behind them.
package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SummitsRepository stores and retrieves summits.
type SummitsRepository interface {
	// UpsertMany inserts summits, updating the ones whose ID already exists.
	UpsertMany(ctx context.Context, summits []Summit) error
}

// Summit mirrors the summits table.
type Summit struct {
	ID         int64             `gorm:"column:id;primaryKey"`
	Names      map[string]string `gorm:"column:names;serializer:json"`
	ElevationM *float32          `gorm:"column:elevation_m"`
	Geom       Point             `gorm:"column:geom"`
}

const upsertBatchSize = 1000

var _ SummitsRepository = (*gormSummits)(nil)

type gormSummits struct {
	db *gorm.DB
}

func NewSummitsRepository(db *gorm.DB) SummitsRepository {
	return &gormSummits{db: db}
}

func (r *gormSummits) UpsertMany(ctx context.Context, summits []Summit) error {
	if len(summits) == 0 {
		return nil
	}
	return gorm.G[Summit](r.db, clause.OnConflict{UpdateAll: true}).
		CreateInBatches(ctx, &summits, upsertBatchSize)
}
