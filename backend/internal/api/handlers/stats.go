package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/anphx/gopackage-chronicle/internal/model"
	repo "github.com/anphx/gopackage-chronicle/internal/repository"
)

const (
	defaultOverviewWeeks = 30
	maxOverviewWeeks     = 104
	defaultActivityWeeks = 12
	maxActivityWeeks     = 52
	maxQueryLength       = 200
)

// StatsHandler serves aggregated views over the index.
type StatsHandler struct {
	statsRepo repo.StatsRepository
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(statsRepo repo.StatsRepository) *StatsHandler {
	return &StatsHandler{statsRepo: statsRepo}
}

// HandleOverview handles GET /api/stats - index totals and weekly release activity.
// Query: weeks (default 30, max 104).
func (h *StatsHandler) HandleOverview(w http.ResponseWriter, r *http.Request) {
	weeks := parseWeeks(r, defaultOverviewWeeks, maxOverviewWeeks)

	overview, err := h.statsRepo.Overview(r.Context(), weeks)
	if err != nil {
		slog.Error("failed to get stats overview", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

// HandleRecentPackages handles GET /api/stats/packages - modules released within the last
// `weeks` weeks, newest first. Query: limit, offset, weeks (default 12, max 52),
// q (path substring), kind (major|minor|patch|prerelease), sort (recent|active|name).
func (h *StatsHandler) HandleRecentPackages(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	weeks := parseWeeks(r, defaultActivityWeeks, maxActivityWeeks)
	params := r.URL.Query()

	kind := params.Get("kind")
	if !isValidKind(kind) {
		http.Error(w, "kind must be one of major, minor, patch, prerelease", http.StatusBadRequest)
		return
	}

	sort := params.Get("sort")
	if sort == "" {
		sort = repo.SortRecent
	}
	if !repo.IsValidSort(sort) {
		http.Error(w, "sort must be one of recent, active, name", http.StatusBadRequest)
		return
	}

	query := strings.TrimSpace(params.Get("q"))
	if len(query) > maxQueryLength {
		http.Error(w, "q is too long", http.StatusBadRequest)
		return
	}

	page, err := h.statsRepo.RecentPackages(r.Context(), repo.RecentPackagesFilter{
		Weeks:  weeks,
		Query:  query,
		Kind:   kind,
		Sort:   sort,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		slog.Error("failed to get recent packages", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"packages":    page.Packages,
		"total":       page.Total,
		"kind_counts": page.KindCounts,
		"limit":       limit,
		"offset":      offset,
		"weeks":       weeks,
	})
}

// parseWeeks reads the "weeks" query parameter, falling back to def for missing or invalid
// values and capping it at maxWeeks.
func parseWeeks(r *http.Request, def, maxWeeks int) int {
	if v := r.URL.Query().Get("weeks"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			return min(parsed, maxWeeks)
		}
	}
	return def
}

func isValidKind(kind string) bool {
	switch kind {
	case "", model.KindMajor, model.KindMinor, model.KindPatch, model.KindPrerelease:
		return true
	default:
		return false
	}
}
