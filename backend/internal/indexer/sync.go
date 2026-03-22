package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/anphx/gopackage-chronicles/internal/model"
	repo "github.com/anphx/gopackage-chronicles/internal/repository"
)

const (
	// TODO: further improvement, config
	// These values can be tuned based on performance testing and the expected load.
	batchSize        = 2000
	maxBatchesPerRun = 20 // Limit per run to avoid long-running jobs
)

// Syncer coordinates fetching releases and storing them in the database.
type Syncer struct {
	client      IndexClient
	packageRepo repo.PackageRepository
	releaseRepo repo.ReleaseRepository
	cursorRepo  repo.SyncCursorRepository
}

// NewSyncer creates a new Syncer.
func NewSyncer(
	client IndexClient,
	packageRepo repo.PackageRepository,
	releaseRepo repo.ReleaseRepository,
	cursorRepo repo.SyncCursorRepository,
) *Syncer {
	return &Syncer{
		client:      client,
		packageRepo: packageRepo,
		releaseRepo: releaseRepo,
		cursorRepo:  cursorRepo,
	}
}

// Run executes the sync process.
func (s *Syncer) Run(ctx context.Context) error {
	lastSynced, err := s.cursorRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("syncer: get cursor: %w", err)
	}

	if lastSynced.IsZero() {
		// Start from 1 year ago if this is the first run to collect sufficient historical data,
		// but not too far back, that data size becomes an issue.
		lastSynced = time.Now().Add(-24 * time.Hour * 365)
		slog.Info("first sync run, starting from", "since", lastSynced)
	}

	totalPackages := 0
	totalReleases := 0
	batchCount := 0

	for batchCount < maxBatchesPerRun {
		entries, err := s.client.FetchReleases(lastSynced, batchSize)
		if err != nil {
			return fmt.Errorf("syncer: fetch releases: %w", err)
		}

		if len(entries) == 0 {
			slog.Info("no new releases found")
			break
		}

		slog.Info(fmt.Sprintf("fetched entries - %d", batchCount+1), "count", len(entries))

		// Process entries
		packagesProcessed, releasesProcessed, err := s.processEntries(ctx, entries)
		if err != nil {
			return fmt.Errorf("syncer: process entries: %w", err)
		}

		totalPackages += packagesProcessed
		totalReleases += releasesProcessed

		// Update the cursor to the timestamp of the last entry
		lastTimestamp := entries[len(entries)-1].Timestamp
		if err := s.cursorRepo.Update(ctx, lastTimestamp); err != nil {
			return fmt.Errorf("syncer: update cursor: %w", err)
		}

		lastSynced = lastTimestamp
		batchCount++

		// If we got fewer than batchSize, we've caught up
		if len(entries) < batchSize {
			slog.Info("caught up with index")
			break
		}
	}

	slog.Info("sync completed",
		"total_packages", totalPackages,
		"total_releases", totalReleases,
		"batches", batchCount,
		"lastSyncedAt", lastSynced,
	)

	return nil
}

func (s *Syncer) processEntries(ctx context.Context, entries []IndexEntry) (int, int, error) {
	packagesMap := make(map[string]*model.Package)
	var releases []model.Release

	// Step 1 - create/get all packages
	for _, entry := range entries {
		if _, exists := packagesMap[entry.Path]; !exists {
			pkg, err := s.packageRepo.Create(ctx, entry.Path)
			if err != nil {
				return 0, 0, fmt.Errorf("syncer: create package %s: %w", entry.Path, err)
			}
			packagesMap[entry.Path] = pkg
		}
	}

	// Step 2 - prepare releases per package
	for _, entry := range entries {
		pkg := packagesMap[entry.Path]
		releases = append(releases, model.Release{
			PackageID:  pkg.ID,
			Version:    entry.Version,
			ReleasedAt: entry.Timestamp,
		})
	}

	// Batch insert releases
	if err := s.releaseRepo.CreateBatch(ctx, releases); err != nil {
		return 0, 0, fmt.Errorf("syncer: create releases: %w", err)
	}

	return len(packagesMap), len(releases), nil
}
