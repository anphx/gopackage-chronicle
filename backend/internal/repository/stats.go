package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anphx/gopackage-chronicle/internal/model"
)

// Sort orders accepted by RecentPackages.
const (
	SortRecent = "recent" // newest release first (default)
	SortActive = "active" // most releases inside the window first
	SortName   = "name"   // module path, A to Z
)

// recentPackagesOrder returns the ORDER BY clause for a sort option. Only these fixed
// strings are ever placed in the query.
func recentPackagesOrder(sort string) (string, bool) {
	switch sort {
	case SortRecent:
		return "rc.last_released_at DESC, rc.package_id DESC", true
	case SortActive:
		return "rc.window_count DESC, rc.last_released_at DESC, rc.package_id DESC", true
	case SortName:
		return "rc.path ASC", true
	default:
		return "", false
	}
}

// IsValidSort reports whether sort is a supported RecentPackages sort order.
func IsValidSort(sort string) bool {
	_, ok := recentPackagesOrder(sort)
	return ok
}

// versionKindSQL classifies a version string by its shape:
//   - anything with a pre-release part is a prerelease (pseudo-versions are not stored),
//   - vX.0.0 is a major release, vX.Y.0 a minor release, anything else a patch.
//
// Build metadata such as "+incompatible" is ignored. %[1]s is the column to classify.
const versionKindSQL = `CASE
	WHEN %[1]s ~ '^[^+]*-' THEN 'prerelease'
	WHEN %[1]s ~ '^v[0-9]+\.0\.0(\+.*)?$' THEN 'major'
	WHEN %[1]s ~ '^v[0-9]+\.[0-9]+\.0(\+.*)?$' THEN 'minor'
	ELSE 'patch'
END`

// RecentPackagesFilter selects a page of recently updated modules.
type RecentPackagesFilter struct {
	Weeks  int    // window size in weeks, ending with the current UTC week
	Query  string // case-insensitive substring of the module path; empty matches all
	Kind   string // kind of the latest release (model.Kind*); empty matches all
	Sort   string // one of the Sort* constants; empty means SortRecent
	Limit  int
	Offset int
}

// StatsRepository provides aggregated, read-only views over releases.
type StatsRepository interface {
	Overview(ctx context.Context, weeks int) (*model.StatsOverview, error)
	RecentPackages(ctx context.Context, filter RecentPackagesFilter) (*model.RecentPackagesPage, error)
}

// StatsRepo implements StatsRepository using PostgreSQL.
type StatsRepo struct {
	db *sql.DB
}

// NewStatsRepository creates a new StatsRepository.
func NewStatsRepository(db *sql.DB) StatsRepository {
	return &StatsRepo{db: db}
}

// Overview returns index-wide totals and per-week release counts for the last `weeks` UTC weeks,
// the current (partial) week included. Weeks without releases are returned with zero counts.
func (r *StatsRepo) Overview(ctx context.Context, weeks int) (*model.StatsOverview, error) {
	overview := &model.StatsOverview{Weeks: make([]model.WeeklyActivity, 0, weeks)}

	var lastSynced sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT (SELECT count(*) FROM releases),
		       (SELECT count(*) FROM packages),
		       (SELECT last_synced FROM sync_cursor WHERE id = 1)
	`).Scan(&overview.TotalReleases, &overview.TotalPackages, &lastSynced)
	if err != nil {
		return nil, fmt.Errorf("repository: query totals: %w", err)
	}
	if lastSynced.Valid {
		overview.LastSyncedAt = &lastSynced.Time
	}

	query := fmt.Sprintf(`
		WITH b AS (
			SELECT date_trunc('week', now() AT TIME ZONE 'UTC') - make_interval(weeks => $1 - 1) AS since
		), agg AS (
			SELECT date_trunc('week', r.released_at AT TIME ZONE 'UTC') AS week_start,
			       %s AS kind
			FROM releases r, b
			WHERE r.released_at >= b.since AT TIME ZONE 'UTC'
		)
		SELECT gs.week_start AT TIME ZONE 'UTC',
		       count(agg.kind) FILTER (WHERE agg.kind = 'major'),
		       count(agg.kind) FILTER (WHERE agg.kind = 'minor'),
		       count(agg.kind) FILTER (WHERE agg.kind = 'patch'),
		       count(agg.kind) FILTER (WHERE agg.kind = 'prerelease')
		FROM b, generate_series(b.since, b.since + make_interval(weeks => $1 - 1), interval '1 week') AS gs(week_start)
		LEFT JOIN agg ON agg.week_start = gs.week_start
		GROUP BY gs.week_start
		ORDER BY gs.week_start
	`, fmt.Sprintf(versionKindSQL, "r.version"))

	rows, err := r.db.QueryContext(ctx, query, weeks)
	if err != nil {
		return nil, fmt.Errorf("repository: query weekly activity: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var w model.WeeklyActivity
		if err := rows.Scan(&w.WeekStart, &w.Major, &w.Minor, &w.Patch, &w.Prerelease); err != nil {
			return nil, fmt.Errorf("repository: scan weekly activity: %w", err)
		}
		overview.Weeks = append(overview.Weeks, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows error: %w", err)
	}

	return overview, nil
}

// recentPackagesCTE selects, for every module released within the window, its latest release.
// Parameters: $1 window in weeks, $2 ILIKE pattern for the path (” for no filter).
func recentPackagesCTE() string {
	return fmt.Sprintf(`
	WITH b AS (
		SELECT (date_trunc('week', now() AT TIME ZONE 'UTC') - make_interval(weeks => $1 - 1)) AT TIME ZONE 'UTC' AS since
	), win AS (
		SELECT r.package_id, r.version, r.released_at,
		       count(*) OVER (PARTITION BY r.package_id) AS window_count,
		       row_number() OVER (PARTITION BY r.package_id ORDER BY r.released_at DESC, r.version DESC) AS rn
		FROM releases r, b
		WHERE r.released_at >= b.since
	), recent AS (
		SELECT w.package_id, p.path, w.version AS latest_version, w.released_at AS last_released_at,
		       w.window_count, %s AS kind
		FROM win w
		JOIN packages p ON p.id = w.package_id
		WHERE w.rn = 1 AND ($2 = '' OR p.path ILIKE $2 ESCAPE '\')
	)`, fmt.Sprintf(versionKindSQL, "w.version"))
}

// RecentPackages returns one page of modules released within the window, with per-kind counts
// of the whole (query-filtered) result for the filter chips.
func (r *StatsRepo) RecentPackages(ctx context.Context, f RecentPackagesFilter) (*model.RecentPackagesPage, error) {
	order, ok := recentPackagesOrder(f.Sort)
	if !ok {
		order, _ = recentPackagesOrder(SortRecent)
	}
	cte := recentPackagesCTE()
	pattern := likePattern(f.Query)

	page := &model.RecentPackagesPage{Packages: make([]model.PackageActivity, 0, f.Limit)}

	err := r.db.QueryRowContext(ctx, cte+`
		SELECT count(*),
		       count(*) FILTER (WHERE kind = 'major'),
		       count(*) FILTER (WHERE kind = 'minor'),
		       count(*) FILTER (WHERE kind = 'patch'),
		       count(*) FILTER (WHERE kind = 'prerelease')
		FROM recent
	`, f.Weeks, pattern).Scan(
		&page.KindCounts.All,
		&page.KindCounts.Major,
		&page.KindCounts.Minor,
		&page.KindCounts.Patch,
		&page.KindCounts.Prerelease,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: count recent packages: %w", err)
	}
	page.Total = kindTotal(page.KindCounts, f.Kind)

	query := cte + `
		SELECT rc.package_id, rc.path, rc.latest_version, rc.kind, rc.last_released_at, rc.window_count,
		       (SELECT count(*) FROM releases x WHERE x.package_id = rc.package_id),
		       (SELECT coalesce(json_agg(v.version ORDER BY v.released_at DESC, v.version DESC), '[]'::json)
		          FROM (SELECT x.version, x.released_at
		                FROM releases x
		                WHERE x.package_id = rc.package_id
		                ORDER BY x.released_at DESC, x.version DESC
		                LIMIT 4) v),
		       (SELECT json_agg(coalesce(c.n, 0) ORDER BY gs.i)
		          FROM generate_series(0, $1 - 1) AS gs(i)
		          LEFT JOIN (SELECT floor(extract(epoch FROM x.released_at - b.since) / 604800)::int AS i,
		                            count(*) AS n
		                     FROM releases x
		                     WHERE x.package_id = rc.package_id AND x.released_at >= b.since
		                     GROUP BY 1) c ON c.i = gs.i)
		FROM recent rc, b
		WHERE ($3 = '' OR rc.kind = $3)
		ORDER BY ` + order + `
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.QueryContext(ctx, query, f.Weeks, pattern, f.Kind, f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("repository: query recent packages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			p              model.PackageActivity
			recentVersions []byte
			weekly         []byte
		)
		if err := rows.Scan(
			&p.PackageID,
			&p.Path,
			&p.LatestVersion,
			&p.LatestKind,
			&p.LastReleasedAt,
			&p.WindowReleases,
			&p.ReleaseCount,
			&recentVersions,
			&weekly,
		); err != nil {
			return nil, fmt.Errorf("repository: scan recent package: %w", err)
		}
		if err := json.Unmarshal(recentVersions, &p.RecentVersions); err != nil {
			return nil, fmt.Errorf("repository: decode recent versions: %w", err)
		}
		if err := json.Unmarshal(weekly, &p.Weekly); err != nil {
			return nil, fmt.Errorf("repository: decode weekly counts: %w", err)
		}
		page.Packages = append(page.Packages, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows error: %w", err)
	}

	return page, nil
}

// likePattern turns a free-text query into an ILIKE substring pattern, escaping LIKE
// wildcards. An empty query yields an empty pattern, which the SQL treats as "no filter".
func likePattern(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}
	escaper := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + escaper.Replace(q) + "%"
}

// likePrefixPattern returns a LIKE pattern for paths starting with q.
func likePrefixPattern(q string) string {
	return strings.TrimPrefix(likePattern(q), "%")
}

// kindTotal picks the count matching a kind filter; an empty kind means all modules.
func kindTotal(c model.KindCounts, kind string) int64 {
	switch kind {
	case model.KindMajor:
		return c.Major
	case model.KindMinor:
		return c.Minor
	case model.KindPatch:
		return c.Patch
	case model.KindPrerelease:
		return c.Prerelease
	default:
		return c.All
	}
}
