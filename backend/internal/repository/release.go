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

// CreateBatch inserts multiple releases using multi-row INSERT for better performance.
// This is 10-100x faster than individual inserts for large batches.
func (r *ReleaseRepo) CreateBatch(ctx context.Context, releases []model.Release) error {
	if len(releases) == 0 {
		return nil
	}

	// Build multi-row INSERT query
	// For very large batches, process in chunks of 1000 to avoid parameter limits
	const batchSize = 1000

	for i := 0; i < len(releases); i += batchSize {
		end := i + batchSize
		if end > len(releases) {
			end = len(releases)
		}

		batch := releases[i:end]
		if err := r.insertBatch(ctx, batch); err != nil {
			return err
		}
	}

	return nil
}

// insertBatch performs a single multi-row INSERT.
func (r *ReleaseRepo) insertBatch(ctx context.Context, releases []model.Release) error {
	if len(releases) == 0 {
		return nil
	}

	// Build VALUES clause with placeholders
	values := make([]any, 0, len(releases)*3)
	placeholders := make([]string, 0, len(releases))

	for i, release := range releases {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		values = append(values, release.PackageID, release.Version, release.ReleasedAt)
	}

	query := fmt.Sprintf(`
		INSERT INTO releases (package_id, version, released_at)
		VALUES %s
		ON CONFLICT (package_id, version) DO NOTHING
	`, joinStrings(placeholders, ", "))

	_, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("repository: insert release batch: %w", err)
	}

	return nil
}

// joinStrings is a simple helper to join strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
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
