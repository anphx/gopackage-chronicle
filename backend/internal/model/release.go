package model

import "time"

// Release represents a tagged version of a package. It is identified by (PackageID, Version).
type Release struct {
	PackageID  int64     `json:"package_id"`
	Version    string    `json:"version"`
	ReleasedAt time.Time `json:"released_at"`
}

// ReleaseWithPackage represents a release joined with its package information.
type ReleaseWithPackage struct {
	Release
	PackagePath string `json:"package_path"`
}
