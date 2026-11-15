package repository

import "testing"

func TestLikePatternsTreatInputAsSearchData(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantPattern string
		wantPrefix  string
	}{
		{
			name:        "SQL-looking input remains a literal pattern",
			query:       "' OR 1=1 --",
			wantPattern: "%' OR 1=1 --%",
			wantPrefix:  "' OR 1=1 --%",
		},
		{
			name:        "escapes LIKE metacharacters",
			query:       `a%_\b`,
			wantPattern: `%a\%\_\\b%`,
			wantPrefix:  `a\%\_\\b%`,
		},
		{
			name:        "empty input disables filtering",
			query:       "  ",
			wantPattern: "",
			wantPrefix:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := likePattern(tt.query); got != tt.wantPattern {
				t.Errorf("likePattern(%q) = %q, want %q", tt.query, got, tt.wantPattern)
			}
			if got := likePrefixPattern(tt.query); got != tt.wantPrefix {
				t.Errorf("likePrefixPattern(%q) = %q, want %q", tt.query, got, tt.wantPrefix)
			}
		})
	}
}

func TestRecentPackagesOrderOnlyAllowsKnownSorts(t *testing.T) {
	tests := []struct {
		name   string
		sort   string
		want   string
		wantOK bool
	}{
		{name: "recent", sort: SortRecent, want: "rc.last_released_at DESC, rc.package_id DESC", wantOK: true},
		{name: "active", sort: SortActive, want: "rc.window_count DESC, rc.last_released_at DESC, rc.package_id DESC", wantOK: true},
		{name: "name", sort: SortName, want: "rc.path ASC", wantOK: true},
		{name: "rejects SQL", sort: "recent; DROP TABLE packages;--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := recentPackagesOrder(tt.sort)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("recentPackagesOrder(%q) = (%q, %v), want (%q, %v)", tt.sort, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
