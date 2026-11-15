package indexer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anphx/gopackage-chronicle/internal/model"
)

type mockClient struct {
	batches [][]IndexEntry
	calls   int
}

func (m *mockClient) FetchReleases(_ time.Time, _ int) ([]IndexEntry, error) {
	if m.calls >= len(m.batches) {
		return nil, nil
	}
	b := m.batches[m.calls]
	m.calls++
	return b, nil
}

type mockPackageRepo struct {
	upsertCalls [][]string
	ids         map[string]int64
	err         error
}

func (m *mockPackageRepo) Create(_ context.Context, _ string) (*model.Package, error) {
	return nil, errors.New("Create must not be called by the syncer")
}
func (m *mockPackageRepo) UpsertBatch(_ context.Context, paths []string) (map[string]int64, error) {
	m.upsertCalls = append(m.upsertCalls, paths)
	if m.err != nil {
		return nil, m.err
	}
	out := make(map[string]int64, len(paths))
	for _, p := range paths {
		if id, ok := m.ids[p]; ok {
			out[p] = id
		}
	}
	return out, nil
}
func (m *mockPackageRepo) GetByPath(_ context.Context, _ string) (*model.Package, error) {
	return nil, nil
}
func (m *mockPackageRepo) List(_ context.Context, _, _ int) ([]model.Package, error) {
	return nil, nil
}
func (m *mockPackageRepo) Search(_ context.Context, _ string, _, _ int) ([]model.Package, error) {
	return nil, nil
}

type mockReleaseRepo struct {
	inserted []model.Release
}

func (m *mockReleaseRepo) CreateBatch(_ context.Context, releases []model.Release) error {
	m.inserted = append(m.inserted, releases...)
	return nil
}
func (m *mockReleaseRepo) GetByPackageID(_ context.Context, _ int64, _, _ int) ([]model.Release, error) {
	return nil, nil
}
func (m *mockReleaseRepo) GetRecent(_ context.Context, _, _ int) ([]model.ReleaseWithPackage, error) {
	return nil, nil
}

type mockCursorRepo struct {
	last time.Time
}

func (m *mockCursorRepo) Get(_ context.Context) (time.Time, error) { return m.last, nil }
func (m *mockCursorRepo) Update(_ context.Context, ts time.Time) error {
	m.last = ts
	return nil
}

func TestSyncerRun(t *testing.T) {
	t1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Minute)
	t3 := t2.Add(time.Minute)

	tests := []struct {
		name            string
		batches         [][]IndexEntry
		ids             map[string]int64
		upsertErr       error
		wantErr         bool
		wantUpserts     [][]string
		wantReleaseIDs  []int64
		wantCursorAfter time.Time
	}{
		{
			name: "one upsert per batch with deduplicated paths",
			batches: [][]IndexEntry{{
				{Path: "a", Version: "v1.0.0", Timestamp: t1},
				{Path: "b", Version: "v0.1.0", Timestamp: t2},
				{Path: "a", Version: "v1.1.0", Timestamp: t3},
			}},
			ids:             map[string]int64{"a": 1, "b": 2},
			wantUpserts:     [][]string{{"a", "b"}},
			wantReleaseIDs:  []int64{1, 2, 1},
			wantCursorAfter: t3,
		},
		{
			name: "pseudo-versions are skipped",
			batches: [][]IndexEntry{{
				{Path: "a", Version: "v1.0.0", Timestamp: t1},
				{Path: "b", Version: "v0.0.0-20260301000100-3f2a9c1d8e7b", Timestamp: t2},
				{Path: "a", Version: "v1.2.4-0.20260301000200-3f2a9c1d8e7b", Timestamp: t3},
				{Path: "c", Version: "v1.2.4-rc.1.0.20260301000200-3f2a9c1d8e7b", Timestamp: t3},
				{Path: "d", Version: "v2.0.0-rc.1", Timestamp: t3},
			}},
			ids:             map[string]int64{"a": 1, "d": 4},
			wantUpserts:     [][]string{{"a", "d"}},
			wantReleaseIDs:  []int64{1, 4},
			wantCursorAfter: t3,
		},
		{
			name: "batch of only pseudo-versions still advances the cursor",
			batches: [][]IndexEntry{{
				{Path: "b", Version: "v0.0.0-20260301000100-3f2a9c1d8e7b", Timestamp: t2},
			}},
			wantUpserts:     nil,
			wantCursorAfter: t2,
		},
		{
			name: "missing id from repository is an error",
			batches: [][]IndexEntry{{
				{Path: "a", Version: "v1.0.0", Timestamp: t1},
			}},
			ids:         map[string]int64{},
			wantErr:     true,
			wantUpserts: [][]string{{"a"}},
		},
		{
			name: "upsert error is returned",
			batches: [][]IndexEntry{{
				{Path: "a", Version: "v1.0.0", Timestamp: t1},
			}},
			upsertErr:   errors.New("boom"),
			wantErr:     true,
			wantUpserts: [][]string{{"a"}},
		},
		{
			name:        "no entries",
			batches:     nil,
			wantUpserts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs := &mockPackageRepo{ids: tt.ids, err: tt.upsertErr}
			rels := &mockReleaseRepo{}
			cursor := &mockCursorRepo{}
			s := NewSyncer(&mockClient{batches: tt.batches}, pkgs, rels, cursor)

			err := s.Run(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("Run() err = %v, wantErr %v", err, tt.wantErr)
			}

			if len(pkgs.upsertCalls) != len(tt.wantUpserts) {
				t.Fatalf("UpsertBatch calls = %v, want %v", pkgs.upsertCalls, tt.wantUpserts)
			}
			for i, call := range pkgs.upsertCalls {
				if len(call) != len(tt.wantUpserts[i]) {
					t.Fatalf("UpsertBatch call %d = %v, want %v", i, call, tt.wantUpserts[i])
				}
				for j := range call {
					if call[j] != tt.wantUpserts[i][j] {
						t.Errorf("UpsertBatch call %d = %v, want %v", i, call, tt.wantUpserts[i])
					}
				}
			}

			if tt.wantErr {
				return
			}

			if len(rels.inserted) != len(tt.wantReleaseIDs) {
				t.Fatalf("inserted %d releases, want %d", len(rels.inserted), len(tt.wantReleaseIDs))
			}
			for i, r := range rels.inserted {
				if r.PackageID != tt.wantReleaseIDs[i] {
					t.Errorf("release %d package id = %d, want %d", i, r.PackageID, tt.wantReleaseIDs[i])
				}
			}
			if !tt.wantCursorAfter.IsZero() && !cursor.last.Equal(tt.wantCursorAfter) {
				t.Errorf("cursor = %v, want %v", cursor.last, tt.wantCursorAfter)
			}
		})
	}
}
