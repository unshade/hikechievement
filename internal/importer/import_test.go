package importer

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/unshade/hikechievement/internal/repository"
)

var _ repository.SummitsRepository = (*fakeRepo)(nil)

// fakeRepo records every UpsertMany call so tests can inspect batching
// without a real database.
type fakeRepo struct {
	batches [][]repository.Summit
	err     error
}

func (r *fakeRepo) UpsertMany(_ context.Context, summits []repository.Summit) error {
	if r.err != nil {
		return r.err
	}
	r.batches = append(r.batches, slices.Clone(summits))
	return nil
}

// peaks builds n distinct peaks (id 1..n), each with a name and an elevation.
func peaks(n int) []Peak {
	out := make([]Peak, n)
	for i := range out {
		ele := float64(1000 + i)
		out[i] = Peak{
			ID:         int64(i + 1),
			Names:      map[string]string{DefaultName: "Peak"},
			ElevationM: &ele,
			Lat:        float64(i),
			Lon:        float64(-i),
		}
	}
	return out
}

func sourceOf(ps []Peak) func(func(Peak) error) error {
	return func(fn func(Peak) error) error {
		for _, p := range ps {
			if err := fn(p); err != nil {
				return err
			}
		}
		return nil
	}
}

func TestBatchImport(t *testing.T) {
	t.Run("no peaks upserts nothing", func(t *testing.T) {
		repo := &fakeRepo{}
		n, err := batchImport(t.Context(), sourceOf(nil), repo, 10)
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 || len(repo.batches) != 0 {
			t.Errorf("n = %d, batches = %v, want 0 and none", n, repo.batches)
		}
	})

	t.Run("fewer peaks than the batch size flush once at the end", func(t *testing.T) {
		repo := &fakeRepo{}
		n, err := batchImport(t.Context(), sourceOf(peaks(3)), repo, 10)
		if err != nil {
			t.Fatal(err)
		}
		if n != 3 {
			t.Errorf("n = %d, want 3", n)
		}
		if len(repo.batches) != 1 || len(repo.batches[0]) != 3 {
			t.Fatalf("batches = %v, want a single batch of 3", repo.batches)
		}
	})

	t.Run("splits peaks into full batches plus a remainder", func(t *testing.T) {
		repo := &fakeRepo{}
		n, err := batchImport(t.Context(), sourceOf(peaks(25)), repo, 10)
		if err != nil {
			t.Fatal(err)
		}
		if n != 25 {
			t.Errorf("n = %d, want 25", n)
		}
		wantSizes := []int{10, 10, 5}
		if len(repo.batches) != len(wantSizes) {
			t.Fatalf("batches = %d, want %d", len(repo.batches), len(wantSizes))
		}
		for i, want := range wantSizes {
			if len(repo.batches[i]) != want {
				t.Errorf("batch %d size = %d, want %d", i, len(repo.batches[i]), want)
			}
		}
	})

	t.Run("a peak exactly filling the batch size does not double flush", func(t *testing.T) {
		repo := &fakeRepo{}
		n, err := batchImport(t.Context(), sourceOf(peaks(10)), repo, 10)
		if err != nil {
			t.Fatal(err)
		}
		if n != 10 || len(repo.batches) != 1 {
			t.Errorf("n = %d, batches = %d, want 10 and 1", n, len(repo.batches))
		}
	})

	t.Run("converts peak fields to a summit", func(t *testing.T) {
		repo := &fakeRepo{}
		ele := 4478.5
		p := Peak{
			ID:         42,
			Names:      map[string]string{DefaultName: "Matterhorn", "fr": "Cervin"},
			ElevationM: &ele,
			Lat:        45.9763,
			Lon:        7.6586,
		}
		if _, err := batchImport(t.Context(), sourceOf([]Peak{p}), repo, 10); err != nil {
			t.Fatal(err)
		}
		got := repo.batches[0][0]
		want := repository.Summit{
			ID:         42,
			Names:      map[string]string{DefaultName: "Matterhorn", "fr": "Cervin"},
			ElevationM: func() *float32 { v := float32(4478.5); return &v }(),
			Geom:       repository.Point{7.6586, 45.9763},
		}
		if got.ID != want.ID || got.Geom != want.Geom || *got.ElevationM != *want.ElevationM {
			t.Errorf("summit = %+v, want %+v", got, want)
		}
		if got.Names[DefaultName] != "Matterhorn" || got.Names["fr"] != "Cervin" {
			t.Errorf("names = %v", got.Names)
		}
	})

	t.Run("a nil elevation stays nil", func(t *testing.T) {
		repo := &fakeRepo{}
		p := Peak{ID: 1, Lat: 1, Lon: 2}
		if _, err := batchImport(t.Context(), sourceOf([]Peak{p}), repo, 10); err != nil {
			t.Fatal(err)
		}
		if got := repo.batches[0][0].ElevationM; got != nil {
			t.Errorf("elevation = %v, want nil", *got)
		}
	})

	t.Run("nil names become an empty map", func(t *testing.T) {
		repo := &fakeRepo{}
		p := Peak{ID: 1, Lat: 1, Lon: 2}
		if _, err := batchImport(t.Context(), sourceOf([]Peak{p}), repo, 10); err != nil {
			t.Fatal(err)
		}
		if got := repo.batches[0][0].Names; got == nil || len(got) != 0 {
			t.Errorf("names = %v, want an empty non-nil map", got)
		}
	})

	t.Run("a repository error stops the import and is returned", func(t *testing.T) {
		wantErr := errors.New("boom")
		repo := &fakeRepo{err: wantErr}
		n, err := batchImport(t.Context(), sourceOf(peaks(3)), repo, 10)
		if !errors.Is(err, wantErr) {
			t.Errorf("err = %v, want %v", err, wantErr)
		}
		// The batch was buffered but the flush at the end failed, so nothing
		// was actually upserted, yet count reflects how many peaks were read.
		if n != 3 {
			t.Errorf("n = %d, want 3", n)
		}
	})

	t.Run("a source error mid-stream stops the import", func(t *testing.T) {
		wantErr := errors.New("bad pbf")
		src := func(fn func(Peak) error) error {
			for _, p := range peaks(3) {
				if err := fn(p); err != nil {
					return err
				}
			}
			return wantErr
		}
		repo := &fakeRepo{}
		n, err := batchImport(t.Context(), src, repo, 10)
		if !errors.Is(err, wantErr) {
			t.Errorf("err = %v, want %v", err, wantErr)
		}
		if n != 3 {
			t.Errorf("n = %d, want 3", n)
		}
	})
}
