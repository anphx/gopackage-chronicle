package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/anphx/gopackage-chronicle/internal/model"
	repo "github.com/anphx/gopackage-chronicle/internal/repository"
)

const (
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
		// Start from Jan 1st 2026.
		lastSynced = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
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
	// Step 1 - create/get all packages in one round trip.
	paths := make([]string, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if _, ok := seen[entry.Path]; !ok {
			seen[entry.Path] = struct{}{}
			paths = append(paths, entry.Path)
		}
	}

	packageIDs, err := s.packageRepo.UpsertBatch(ctx, paths)
	if err != nil {
		return 0, 0, fmt.Errorf("syncer: upsert packages: %w", err)
	}

	// Step 2 - prepare releases per package
	releases := make([]model.Release, 0, len(entries))
	for _, entry := range entries {
		id, ok := packageIDs[entry.Path]
		if !ok {
			return 0, 0, fmt.Errorf("syncer: no package id returned for %s", entry.Path)
		}
		releases = append(releases, model.Release{
			PackageID:  id,
			Version:    entry.Version,
			ReleasedAt: entry.Timestamp,
		})
	}

	// Batch insert releases
	if err := s.releaseRepo.CreateBatch(ctx, releases); err != nil {
		return 0, 0, fmt.Errorf("syncer: create releases: %w", err)
	}

	return len(paths), len(releases), nil
}
