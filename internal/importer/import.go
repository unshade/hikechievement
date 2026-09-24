package importer

import (
	"context"
	"io"

	"github.com/unshade/hikechievement/internal/repository"
)

// Import reads peaks from r and upserts them into repo in batches of batchSize.
// It returns the number of peaks imported.
func Import(ctx context.Context, r io.Reader, repo repository.SummitsRepository, batchSize int) (int, error) {
	return batchImport(ctx, func(fn func(Peak) error) error {
		return ReadPeaks(ctx, r, fn)
	}, repo, batchSize)
}

// batchImport drives the upsert batching independently of where peaks come from,
// so it can be unit tested without a real .osm.pbf file.
func batchImport(ctx context.Context, source func(func(Peak) error) error, repo repository.SummitsRepository, batchSize int) (int, error) {
	batch := make([]repository.Summit, 0, batchSize)
	count := 0

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := repo.UpsertMany(ctx, batch); err != nil {
			return err
		}
		batch = batch[:0]
		return nil
	}

	err := source(func(p Peak) error {
		batch = append(batch, peakToSummit(p))
		count++
		if len(batch) >= batchSize {
			return flush()
		}
		return nil
	})
	if err != nil {
		return count, err
	}
	if err := flush(); err != nil {
		return count, err
	}
	return count, nil
}

func peakToSummit(p Peak) repository.Summit {
	names := p.Names
	if names == nil {
		names = map[string]string{}
	}

	var ele *float32
	if p.ElevationM != nil {
		v := float32(*p.ElevationM)
		ele = &v
	}

	return repository.Summit{
		ID:         p.ID,
		Names:      names,
		ElevationM: ele,
		Geom:       repository.Point{p.Lon, p.Lat},
	}
}
