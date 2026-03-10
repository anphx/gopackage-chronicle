package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/anphx/gopackage-chronicles/internal/model"
)

// ReleaseRepository defines operations for release data access.
type ReleaseRepository interface {
	CreateBatch(ctx context.Context, releases []model.Release) error
	GetByPackageID(ctx context.Context, packageID int64, limit, offset int) ([]model.Release, error)
	GetRecent(ctx context.Context, limit, offset int) ([]model.ReleaseWithPackage, error)
}

// ReleaseRepo implements ReleaseRepository using PostgreSQL.
type ReleaseRepo struct {
	db *sql.DB
}

// NewReleaseRepository creates a new ReleaseRepository.
func NewReleaseRepository(db *sql.DB) ReleaseRepository {
	return &ReleaseRepo{db: db}
}

// CreateBatch inserts multiple releases, ignoring duplicates.
func (r *ReleaseRepo) CreateBatch(ctx context.Context, releases []model.Release) error {
	if len(releases) == 0 {
		return nil
	}

	query := `
		INSERT INTO releases (package_id, version, released_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (package_id, version) DO NOTHING
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repository: begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("repository: prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, release := range releases {
		if _, err := stmt.ExecContext(ctx, release.PackageID, release.Version, release.ReleasedAt); err != nil {
			return fmt.Errorf("repository: insert release: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository: commit transaction: %w", err)
	}

	return nil
}

// GetByPackageID retrieves releases for a specific package with pagination.
func (r *ReleaseRepo) GetByPackageID(ctx context.Context, packageID int64, limit, offset int) ([]model.Release, error) {
	query := `
		SELECT id, package_id, version, released_at, indexed_at
		FROM releases
		WHERE package_id = $1
		ORDER BY released_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, packageID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: query releases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var releases []model.Release
	for rows.Next() {
		var rel model.Release
		if err := rows.Scan(&rel.ID, &rel.PackageID, &rel.Version, &rel.ReleasedAt, &rel.IndexedAt); err != nil {
			return nil, fmt.Errorf("repository: scan release: %w", err)
		}

		releases = append(releases, rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows error: %w", err)
	}

	return releases, nil
}

// GetRecent retrieves recent releases with package information.
func (r *ReleaseRepo) GetRecent(ctx context.Context, limit, offset int) ([]model.ReleaseWithPackage, error) {
	query := `
		SELECT r.id, r.package_id, r.version, r.released_at, r.indexed_at, p.path
		FROM releases r
		INNER JOIN packages p ON r.package_id = p.id
		ORDER BY r.released_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: query recent releases: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var releases []model.ReleaseWithPackage
	for rows.Next() {
		var rel model.ReleaseWithPackage
		if err := rows.Scan(
			&rel.ID,
			&rel.PackageID,
			&rel.Version,
			&rel.ReleasedAt,
			&rel.IndexedAt,
			&rel.PackagePath,
		); err != nil {
			return nil, fmt.Errorf("repository: scan release with package: %w", err)
		}
		releases = append(releases, rel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows error: %w", err)
	}

	return releases, nil
}
