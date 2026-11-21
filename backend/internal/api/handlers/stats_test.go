package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anphx/gopackage-chronicle/internal/model"
	repo "github.com/anphx/gopackage-chronicle/internal/repository"
)

type mockStatsRepo struct {
	overviewFn       func(ctx context.Context, weeks int) (*model.StatsOverview, error)
	recentPackagesFn func(ctx context.Context, filter repo.RecentPackagesFilter) (*model.RecentPackagesPage, error)
}

func (m *mockStatsRepo) Overview(ctx context.Context, weeks int) (*model.StatsOverview, error) {
	return m.overviewFn(ctx, weeks)
}

func (m *mockStatsRepo) RecentPackages(ctx context.Context, filter repo.RecentPackagesFilter) (*model.RecentPackagesPage, error) {
	return m.recentPackagesFn(ctx, filter)
}

func TestParseWeeks(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "default when missing", query: "", want: 30},
		{name: "custom value", query: "weeks=8", want: 8},
		{name: "capped at max", query: "weeks=500", want: 104},
		{name: "zero falls back to default", query: "weeks=0", want: 30},
		{name: "negative falls back to default", query: "weeks=-3", want: 30},
		{name: "non-numeric falls back to default", query: "weeks=many", want: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			if got := parseWeeks(r, defaultOverviewWeeks, maxOverviewWeeks); got != tt.want {
				t.Errorf("parseWeeks: got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestHandleOverview(t *testing.T) {
	synced := time.Date(2026, 11, 14, 18, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		query         string
		repoResult    *model.StatsOverview
		repoErr       error
		wantStatus    int
		wantWeeksArg  int
		wantWeekCount int
	}{
		{
			name:  "returns overview with default window",
			query: "",
			repoResult: &model.StatsOverview{
				TotalReleases: 1200,
				TotalPackages: 90,
				LastSyncedAt:  &synced,
				Weeks: []model.WeeklyActivity{
					{WeekStart: synced, Major: 1, Minor: 2, Patch: 3, Prerelease: 4},
				},
			},
			wantStatus:    http.StatusOK,
			wantWeeksArg:  30,
			wantWeekCount: 1,
		},
		{
			name:          "passes custom window to repository",
			query:         "weeks=12",
			repoResult:    &model.StatsOverview{Weeks: []model.WeeklyActivity{}},
			wantStatus:    http.StatusOK,
			wantWeeksArg:  12,
			wantWeekCount: 0,
		},
		{
			name:         "returns 500 on repo error",
			query:        "",
			repoErr:      errors.New("db connection lost"),
			wantStatus:   http.StatusInternalServerError,
			wantWeeksArg: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotWeeks int
			h := NewStatsHandler(&mockStatsRepo{
				overviewFn: func(_ context.Context, weeks int) (*model.StatsOverview, error) {
					gotWeeks = weeks
					return tt.repoResult, tt.repoErr
				},
			})

			r := httptest.NewRequest(http.MethodGet, "/api/stats?"+tt.query, nil)
			w := httptest.NewRecorder()
			h.HandleOverview(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK && w.Header().Get("Cache-Control") != "public, max-age=3600, stale-while-revalidate=3600" {
				t.Errorf("cache-control: got %q", w.Header().Get("Cache-Control"))
			}
			if gotWeeks != tt.wantWeeksArg {
				t.Errorf("weeks passed to repo: got %d, want %d", gotWeeks, tt.wantWeeksArg)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var body model.StatsOverview
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.TotalReleases != tt.repoResult.TotalReleases {
				t.Errorf("total_releases: got %d, want %d", body.TotalReleases, tt.repoResult.TotalReleases)
			}
			if len(body.Weeks) != tt.wantWeekCount {
				t.Errorf("weeks: got %d, want %d", len(body.Weeks), tt.wantWeekCount)
			}
		})
	}
}

func TestHandleRecentPackages(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	page := &model.RecentPackagesPage{
		Packages: []model.PackageActivity{
			{
				PackageID:      1,
				Path:           "github.com/spf13/cobra",
				LatestVersion:  "v1.11.0",
				LatestKind:     model.KindMinor,
				LastReleasedAt: now,
				WindowReleases: 2,
				ReleaseCount:   40,
				RecentVersions: []string{"v1.11.0", "v1.10.2"},
				Weekly:         []int64{0, 1, 0, 1},
			},
		},
		Total:      1,
		KindCounts: model.KindCounts{All: 3, Minor: 1, Patch: 2},
	}

	tests := []struct {
		name       string
		query      string
		repoErr    error
		wantStatus int
		wantFilter repo.RecentPackagesFilter
		wantCalled bool
	}{
		{
			name:       "defaults",
			query:      "",
			wantStatus: http.StatusOK,
			wantFilter: repo.RecentPackagesFilter{Weeks: 12, Sort: repo.SortRecent, Limit: 50},
			wantCalled: true,
		},
		{
			name:       "passes filters, sort and pagination",
			query:      "q=+cobra+&kind=minor&sort=active&limit=25&offset=50&weeks=8",
			wantStatus: http.StatusOK,
			wantFilter: repo.RecentPackagesFilter{Weeks: 8, Query: "cobra", Kind: model.KindMinor, Sort: repo.SortActive, Limit: 25, Offset: 50},
			wantCalled: true,
		},
		{
			name:       "caps weeks at 52",
			query:      "weeks=400",
			wantStatus: http.StatusOK,
			wantFilter: repo.RecentPackagesFilter{Weeks: 52, Sort: repo.SortRecent, Limit: 50},
			wantCalled: true,
		},
		{
			name:       "rejects unknown kind",
			query:      "kind=breaking",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects unknown sort",
			query:      "sort=popular",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects overly long query",
			query:      "q=" + longQuery(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 500 on repo error",
			query:      "",
			repoErr:    errors.New("db connection lost"),
			wantStatus: http.StatusInternalServerError,
			wantFilter: repo.RecentPackagesFilter{Weeks: 12, Sort: repo.SortRecent, Limit: 50},
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				called    bool
				gotFilter repo.RecentPackagesFilter
			)
			h := NewStatsHandler(&mockStatsRepo{
				recentPackagesFn: func(_ context.Context, f repo.RecentPackagesFilter) (*model.RecentPackagesPage, error) {
					called = true
					gotFilter = f
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return page, nil
				},
			})

			r := httptest.NewRequest(http.MethodGet, "/api/stats/packages?"+tt.query, nil)
			w := httptest.NewRecorder()
			h.HandleRecentPackages(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK && w.Header().Get("Cache-Control") != "public, max-age=3600, stale-while-revalidate=3600" {
				t.Errorf("cache-control: got %q", w.Header().Get("Cache-Control"))
			}
			if called != tt.wantCalled {
				t.Fatalf("repo called: got %v, want %v", called, tt.wantCalled)
			}
			if called && gotFilter != tt.wantFilter {
				t.Errorf("filter: got %+v, want %+v", gotFilter, tt.wantFilter)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var body struct {
				Packages   []model.PackageActivity `json:"packages"`
				Total      int64                   `json:"total"`
				KindCounts model.KindCounts        `json:"kind_counts"`
				Limit      int                     `json:"limit"`
				Offset     int                     `json:"offset"`
				Weeks      int                     `json:"weeks"`
			}
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(body.Packages) != 1 || body.Total != 1 || body.KindCounts.All != 3 {
				t.Errorf("unexpected body: %+v", body)
			}
			if body.Limit != tt.wantFilter.Limit || body.Offset != tt.wantFilter.Offset || body.Weeks != tt.wantFilter.Weeks {
				t.Errorf("echoed paging: got limit=%d offset=%d weeks=%d", body.Limit, body.Offset, body.Weeks)
			}
		})
	}
}

func longQuery() string {
	b := make([]byte, maxQueryLength+1)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}
