package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/anphx/gopackage-chronicle/internal/model"
)

var ErrNotFound = errors.New("repository: not found")

// PackageRepository defines operations for package data access.
type PackageRepository interface {
	Create(ctx context.Context, path string) (*model.Package, error)
	UpsertBatch(ctx context.Context, paths []string) (map[string]int64, error)
	GetByPath(ctx context.Context, path string) (*model.Package, error)
	List(ctx context.Context, limit, offset int) ([]model.Package, error)
	Search(ctx context.Context, query string, limit, offset int) ([]model.Package, error)
}

// PackageRepo implements PackageRepository using PostgreSQL.
type PackageRepo struct {
	db *sql.DB
}

// NewPackageRepository creates a new PackageRepository.
func NewPackageRepository(db *sql.DB) PackageRepository {
	return &PackageRepo{db: db}
}

// Create inserts or returns an existing package by path.
func (r *PackageRepo) Create(ctx context.Context, path string) (*model.Package, error) {
	query := `
		INSERT INTO packages (path)
		VALUES ($1)
		ON CONFLICT (path) DO UPDATE SET path = EXCLUDED.path
		RETURNING id, path
	`

	var pkg model.Package
	err := r.db.QueryRowContext(ctx, query, path).Scan(
		&pkg.ID,
		&pkg.Path,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create package: %w", err)
	}

	return &pkg, nil
}

// UpsertBatch inserts any missing packages and returns the ID of every
// requested path, in a single round trip. Paths that already exist are not
// rewritten, so repeated runs leave no dead tuples behind.
func (r *PackageRepo) UpsertBatch(ctx context.Context, paths []string) (map[string]int64, error) {
	ids := make(map[string]int64, len(paths))
	if len(paths) == 0 {
		return ids, nil
	}

	// Rows inserted by the CTE are not visible to the second SELECT (same
	// snapshot), so the UNION ALL never returns a path twice. Existing paths are
	// filtered out before the INSERT: ON CONFLICT alone would still draw an id
	// from the sequence for every conflicting row, and ids are 32-bit.
	query := `
		WITH input AS (
			SELECT DISTINCT unnest($1::text[]) AS path
		), ins AS (
			INSERT INTO packages (path)
			SELECT i.path FROM input i
			WHERE NOT EXISTS (SELECT 1 FROM packages p WHERE p.path = i.path)
			ON CONFLICT (path) DO NOTHING
			RETURNING id, path
		)
		SELECT id, path FROM ins
		UNION ALL
		SELECT p.id, p.path FROM packages p JOIN input i ON i.path = p.path
	`

	rows, err := r.db.QueryContext(ctx, query, paths)
	if err != nil {
		return nil, fmt.Errorf("repository: upsert packages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			id   int64
			path string
		)
		if err := rows.Scan(&id, &path); err != nil {
			return nil, fmt.Errorf("repository: scan upserted package: %w", err)
		}
		ids[path] = id
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows error: %w", err)
	}

	return ids, nil
}

// GetByPath retrieves a package by its path.
func (r *PackageRepo) GetByPath(ctx context.Context, path string) (*model.Package, error) {
	query := `SELECT id, path FROM packages WHERE path = $1`

	var pkg model.Package
	err := r.db.QueryRowContext(ctx, query, path).Scan(
		&pkg.ID,
		&pkg.Path,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: get package by path: %w", err)
	}

	return &pkg, nil
}

// List retrieves packages with pagination.
func (r *PackageRepo) List(ctx context.Context, limit, offset int) ([]model.Package, error) {
	query := `
		SELECT id, path
		FROM packages
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list packages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var packages []model.Package
	for rows.Next() {
		var pkg model.Package
		if err := rows.Scan(&pkg.ID, &pkg.Path); err != nil {
			return nil, fmt.Errorf("repository: scan package: %w", err)
		}
		packages = append(packages, pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows error: %w", err)
	}

	return packages, nil
}

// Search retrieves packages whose paths contain the query, with prefix matches first.
func (r *PackageRepo) Search(ctx context.Context, query string, limit, offset int) ([]model.Package, error) {
	pattern := likePattern(query)
	prefixPattern := likePrefixPattern(query)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, path
		FROM packages
		WHERE path ILIKE $1 ESCAPE '\'
		ORDER BY CASE WHEN path ILIKE $2 ESCAPE '\' THEN 0 ELSE 1 END, path ASC
		LIMIT $3 OFFSET $4
	`, pattern, prefixPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: search packages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var packages []model.Package
	for rows.Next() {
		var pkg model.Package
		if err := rows.Scan(&pkg.ID, &pkg.Path); err != nil {
			return nil, fmt.Errorf("repository: scan searched package: %w", err)
		}
		packages = append(packages, pkg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows error: %w", err)
	}

	return packages, nil
}
