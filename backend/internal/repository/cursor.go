package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SyncCursorRepository defines operations for sync cursor management.
type SyncCursorRepository interface {
	Get(ctx context.Context) (time.Time, error)
	Update(ctx context.Context, timestamp time.Time) error
}

// SyncCursorRepo implements SyncCursorRepository using PostgreSQL.
type SyncCursorRepo struct {
	db *sql.DB
}

// NewSyncCursorRepository creates a new SyncCursorRepository.
func NewSyncCursorRepository(db *sql.DB) SyncCursorRepository {
	return &SyncCursorRepo{db: db}
}

// Get retrieves the last sync timestamp.
func (r *SyncCursorRepo) Get(ctx context.Context) (time.Time, error) {
	query := `SELECT last_synced FROM sync_cursor WHERE id = 1`

	var lastSynced time.Time
	err := r.db.QueryRowContext(ctx, query).Scan(&lastSynced)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return zero time if no cursor exists yet
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("repository: get sync cursor: %w", err)
	}

	return lastSynced, nil
}

// Update sets the last sync timestamp.
func (r *SyncCursorRepo) Update(ctx context.Context, timestamp time.Time) error {
	query := `
		INSERT INTO sync_cursor (id, last_synced)
		VALUES (1, $1)
		ON CONFLICT (id) DO UPDATE SET last_synced = EXCLUDED.last_synced
	`

	if _, err := r.db.ExecContext(ctx, query, timestamp); err != nil {
		return fmt.Errorf("repository: update sync cursor: %w", err)
	}

	return nil
}
