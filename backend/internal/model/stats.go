package model

import "time"

// Release kinds, derived from the shape of a version string (see repository.versionKindSQL).
const (
	KindMajor      = "major"
	KindMinor      = "minor"
	KindPatch      = "patch"
	KindPrerelease = "prerelease"
)

// WeeklyActivity counts the releases published in one UTC week (Monday start), split by kind.
type WeeklyActivity struct {
	WeekStart  time.Time `json:"week_start"`
	Major      int64     `json:"major"`
	Minor      int64     `json:"minor"`
	Patch      int64     `json:"patch"`
	Prerelease int64     `json:"prerelease"`
}

// StatsOverview summarises the whole index for the home page.
type StatsOverview struct {
	TotalReleases int64            `json:"total_releases"`
	TotalPackages int64            `json:"total_packages"`
	LastSyncedAt  *time.Time       `json:"last_synced_at"`
	Weeks         []WeeklyActivity `json:"weeks"`
}

// PackageActivity is one module in the "recently updated" list.
type PackageActivity struct {
	PackageID      int64     `json:"package_id"`
	Path           string    `json:"path"`
	LatestVersion  string    `json:"latest_version"`
	LatestKind     string    `json:"latest_kind"`
	LastReleasedAt time.Time `json:"last_released_at"`
	// WindowReleases is the number of releases inside the requested window of weeks.
	WindowReleases int64 `json:"window_releases"`
	// ReleaseCount is the number of releases of the module, all time.
	ReleaseCount int64 `json:"release_count"`
	// RecentVersions holds up to four newest versions, newest first (the latest included).
	RecentVersions []string `json:"recent_versions"`
	// Weekly holds release counts per week of the window, oldest week first.
	Weekly []int64 `json:"weekly"`
}

// KindCounts counts modules by the kind of their latest release.
type KindCounts struct {
	All        int64 `json:"all"`
	Major      int64 `json:"major"`
	Minor      int64 `json:"minor"`
	Patch      int64 `json:"patch"`
	Prerelease int64 `json:"prerelease"`
}

// RecentPackagesPage is one page of recently updated modules.
type RecentPackagesPage struct {
	Packages []PackageActivity `json:"packages"`
	// Total is the number of modules matching the whole filter, across all pages.
	Total      int64      `json:"total"`
	KindCounts KindCounts `json:"kind_counts"`
}
