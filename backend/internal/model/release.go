package model

import "time"

// Release represents a specific version release of a package.
type Release struct {
	ID         int64     `json:"id"`
	PackageID  int64     `json:"package_id"`
	Version    string    `json:"version"`
	ReleasedAt time.Time `json:"released_at"`
	IndexedAt  time.Time `json:"indexed_at"`
}

// ReleaseWithPackage represents a release joined with its package information.
type ReleaseWithPackage struct {
	Release
	PackagePath string `json:"package_path"`
}
