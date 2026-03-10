package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/anphx/gopackage-chronicles/internal/model"
)

var ErrNotFound = errors.New("repository: not found")

// PackageRepository defines operations for package data access.
type PackageRepository interface {
	Create(ctx context.Context, path string) (*model.Package, error)
	GetByPath(ctx context.Context, path string) (*model.Package, error)
	List(ctx context.Context, limit, offset int) ([]model.Package, error)
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
		RETURNING id, path, created_at
	`

	var pkg model.Package
	err := r.db.QueryRowContext(ctx, query, path).Scan(
		&pkg.ID,
		&pkg.Path,
		&pkg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create package: %w", err)
	}

	return &pkg, nil
}

// GetByPath retrieves a package by its path.
func (r *PackageRepo) GetByPath(ctx context.Context, path string) (*model.Package, error) {
	query := `SELECT id, path, created_at FROM packages WHERE path = $1`

	var pkg model.Package
	err := r.db.QueryRowContext(ctx, query, path).Scan(
		&pkg.ID,
		&pkg.Path,
		&pkg.CreatedAt,
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
		SELECT id, path, created_at
		FROM packages
		ORDER BY created_at DESC
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
		if err := rows.Scan(&pkg.ID, &pkg.Path, &pkg.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: scan package: %w", err)
		}
		packages = append(packages, pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows error: %w", err)
	}

	return packages, nil
}
