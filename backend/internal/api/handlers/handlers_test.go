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

// --- mocks ---

type mockPackageRepo struct {
	getByPathFn func(ctx context.Context, path string) (*model.Package, error)
	listFn      func(ctx context.Context, limit, offset int) ([]model.Package, error)
	searchFn    func(ctx context.Context, query string, limit, offset int) ([]model.Package, error)
	createFn    func(ctx context.Context, path string) (*model.Package, error)
}

func (m *mockPackageRepo) GetByPath(ctx context.Context, path string) (*model.Package, error) {
	return m.getByPathFn(ctx, path)
}
func (m *mockPackageRepo) List(ctx context.Context, limit, offset int) ([]model.Package, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset)
	}
	return nil, nil
}
func (m *mockPackageRepo) Search(ctx context.Context, query string, limit, offset int) ([]model.Package, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, limit, offset)
	}
	return nil, nil
}
func (m *mockPackageRepo) Create(ctx context.Context, path string) (*model.Package, error) {
	if m.createFn != nil {
		return m.createFn(ctx, path)
	}
	return nil, nil
}
func (m *mockPackageRepo) UpsertBatch(_ context.Context, _ []string) (map[string]int64, error) {
	return nil, nil
}

type mockReleaseRepo struct {
	getByPackageIDFn func(ctx context.Context, packageID int64, limit, offset int) ([]model.Release, error)
	getRecentFn      func(ctx context.Context, limit, offset int) ([]model.ReleaseWithPackage, error)
	createBatchFn    func(ctx context.Context, releases []model.Release) error
}

func (m *mockReleaseRepo) GetByPackageID(ctx context.Context, packageID int64, limit, offset int) ([]model.Release, error) {
	if m.getByPackageIDFn != nil {
		return m.getByPackageIDFn(ctx, packageID, limit, offset)
	}
	return nil, nil
}
func (m *mockReleaseRepo) GetRecent(ctx context.Context, limit, offset int) ([]model.ReleaseWithPackage, error) {
	return m.getRecentFn(ctx, limit, offset)
}
func (m *mockReleaseRepo) CreateBatch(ctx context.Context, releases []model.Release) error {
	if m.createBatchFn != nil {
		return m.createBatchFn(ctx, releases)
	}
	return nil
}

// --- parsePagination tests ---

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantLimit  int
		wantOffset int
	}{
		{
			name:       "defaults when no query params",
			query:      "",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "custom limit and offset",
			query:      "limit=10&offset=20",
			wantLimit:  10,
			wantOffset: 20,
		},
		{
			name:       "limit capped at 200",
			query:      "limit=300",
			wantLimit:  200,
			wantOffset: 0,
		},
		{
			name:       "very large limit capped at 200",
			query:      "limit=1000000",
			wantLimit:  200,
			wantOffset: 0,
		},
		{
			name:       "negative offset ignored",
			query:      "offset=-5",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "non-numeric values ignored",
			query:      "limit=abc&offset=xyz",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "zero limit ignored",
			query:      "limit=0",
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "boundary limit of 200 accepted",
			query:      "limit=200",
			wantLimit:  200,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			gotLimit, gotOffset := parsePagination(r)
			if gotLimit != tt.wantLimit {
				t.Errorf("limit: got %d, want %d", gotLimit, tt.wantLimit)
			}
			if gotOffset != tt.wantOffset {
				t.Errorf("offset: got %d, want %d", gotOffset, tt.wantOffset)
			}
		})
	}
}

func TestHandleListPackages(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantPath   string
	}{
		{name: "lists packages without a search query", query: "limit=5&offset=2", wantStatus: http.StatusOK},
		{name: "searches all packages with a trimmed query", query: "q=+github.com%2Fspf13%2F+&limit=8", wantStatus: http.StatusOK, wantPath: "github.com/spf13/cobra"},
		{name: "rejects overly long query", query: "q=" + longQuery(), wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listCalled, searchCalled bool
			pkgRepo := &mockPackageRepo{
				listFn: func(_ context.Context, limit, offset int) ([]model.Package, error) {
					listCalled = true
					if limit != 5 && tt.wantPath == "" {
						t.Errorf("limit: got %d, want 5", limit)
					}
					if offset != 2 && tt.wantPath == "" {
						t.Errorf("offset: got %d, want 2", offset)
					}
					return []model.Package{{ID: 1, Path: "github.com/example/pkg"}}, nil
				},
				searchFn: func(_ context.Context, query string, limit, offset int) ([]model.Package, error) {
					searchCalled = true
					if query != "github.com/spf13/" || limit != 8 || offset != 0 {
						t.Errorf("search args: got query=%q limit=%d offset=%d", query, limit, offset)
					}
					return []model.Package{{ID: 2, Path: tt.wantPath}}, nil
				},
			}
			h := NewPackagesHandler(pkgRepo, &mockReleaseRepo{})
			r := httptest.NewRequest(http.MethodGet, "/api/packages?"+tt.query, nil)
			w := httptest.NewRecorder()
			h.HandleList(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusBadRequest {
				if listCalled || searchCalled {
					t.Fatal("repository called for invalid query")
				}
				return
			}
			if listCalled != (tt.wantPath == "") || searchCalled != (tt.wantPath != "") {
				t.Errorf("repository calls: list=%v search=%v", listCalled, searchCalled)
			}
			var body struct {
				Packages []model.Package `json:"packages"`
			}
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(body.Packages) != 1 {
				t.Fatalf("packages count: got %d, want 1", len(body.Packages))
			}
			if tt.wantPath != "" && body.Packages[0].Path != tt.wantPath {
				t.Errorf("package path: got %q, want %q", body.Packages[0].Path, tt.wantPath)
			}
		})
	}
}

// --- GetRecentReleases tests ---

func TestGetRecentReleases(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		query      string
		repoResult []model.ReleaseWithPackage
		repoErr    error
		wantStatus int
		wantCount  int // expected len of releases in response (only checked on 200)
		wantLimit  int
		wantOffset int
	}{
		{
			name:  "returns releases successfully",
			query: "",
			repoResult: []model.ReleaseWithPackage{
				{Release: model.Release{Version: "v1.0.0", ReleasedAt: now}, PackagePath: "github.com/foo/bar"},
				{Release: model.Release{Version: "v1.1.0", ReleasedAt: now}, PackagePath: "github.com/baz/qux"},
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "returns empty list",
			query:      "",
			repoResult: []model.ReleaseWithPackage{},
			wantStatus: http.StatusOK,
			wantCount:  0,
			wantLimit:  50,
			wantOffset: 0,
		},
		{
			name:       "propagates pagination params to response",
			query:      "limit=5&offset=10",
			repoResult: []model.ReleaseWithPackage{},
			wantStatus: http.StatusOK,
			wantLimit:  5,
			wantOffset: 10,
		},
		{
			name:       "returns 500 on repo error",
			query:      "",
			repoErr:    errors.New("db connection lost"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseRepo := &mockReleaseRepo{
				getRecentFn: func(_ context.Context, limit, offset int) ([]model.ReleaseWithPackage, error) {
					return tt.repoResult, tt.repoErr
				},
			}
			h := NewReleasesHandler(releaseRepo)

			r := httptest.NewRequest(http.MethodGet, "/api/releases?"+tt.query, nil)
			w := httptest.NewRecorder()
			h.GetRecentReleases(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var body struct {
				Releases []model.ReleaseWithPackage `json:"releases"`
				Limit    int                        `json:"limit"`
				Offset   int                        `json:"offset"`
			}
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if len(body.Releases) != tt.wantCount {
				t.Errorf("releases count: got %d, want %d", len(body.Releases), tt.wantCount)
			}
			if body.Limit != tt.wantLimit {
				t.Errorf("limit: got %d, want %d", body.Limit, tt.wantLimit)
			}
			if body.Offset != tt.wantOffset {
				t.Errorf("offset: got %d, want %d", body.Offset, tt.wantOffset)
			}
		})
	}
}

// --- HandleDetail tests ---

func TestHandleDetail(t *testing.T) {
	now := time.Now()
	stubPkg := &model.Package{ID: 42, Path: "github.com/foo/bar"}
	stubReleases := []model.Release{
		{PackageID: 42, Version: "v1.0.0", ReleasedAt: now},
	}

	tests := []struct {
		name            string
		pathValue       string
		getByPathResult *model.Package
		getByPathErr    error
		releasesResult  []model.Release
		releasesErr     error
		wantStatus      int
	}{
		{
			name:            "returns package and releases",
			pathValue:       "github.com/foo/bar",
			getByPathResult: stubPkg,
			releasesResult:  stubReleases,
			wantStatus:      http.StatusOK,
		},
		{
			name:         "returns 404 when package not found",
			pathValue:    "github.com/does/not-exist",
			getByPathErr: repo.ErrNotFound,
			wantStatus:   http.StatusNotFound,
		},
		{
			name:         "returns 500 on unexpected package repo error",
			pathValue:    "github.com/foo/bar",
			getByPathErr: errors.New("unexpected db error"),
			wantStatus:   http.StatusInternalServerError,
		},
		{
			name:            "returns 500 when releases query fails",
			pathValue:       "github.com/foo/bar",
			getByPathResult: stubPkg,
			releasesErr:     errors.New("releases query failed"),
			wantStatus:      http.StatusInternalServerError,
		},
		{
			name:       "returns 400 when name is empty",
			pathValue:  "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgRepo := &mockPackageRepo{
				getByPathFn: func(_ context.Context, path string) (*model.Package, error) {
					return tt.getByPathResult, tt.getByPathErr
				},
			}
			releaseRepo := &mockReleaseRepo{
				getByPackageIDFn: func(_ context.Context, _ int64, _, _ int) ([]model.Release, error) {
					return tt.releasesResult, tt.releasesErr
				},
			}
			h := NewPackagesHandler(pkgRepo, releaseRepo)

			r := httptest.NewRequest(http.MethodGet, "/api/packages/"+tt.pathValue, nil)
			// Simulate Go 1.22 routing path value
			if tt.pathValue != "" {
				r.SetPathValue("name", tt.pathValue)
			}
			w := httptest.NewRecorder()
			h.HandleDetail(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var body struct {
				Package  model.Package   `json:"package"`
				Releases []model.Release `json:"releases"`
			}
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if body.Package.Path != stubPkg.Path {
				t.Errorf("package path: got %q, want %q", body.Package.Path, stubPkg.Path)
			}
			if len(body.Releases) != len(tt.releasesResult) {
				t.Errorf("releases count: got %d, want %d", len(body.Releases), len(tt.releasesResult))
			}
		})
	}
}
