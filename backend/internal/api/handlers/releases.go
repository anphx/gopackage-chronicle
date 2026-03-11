package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	repo "github.com/anphx/gopackage-chronicles/internal/repository"
)

const (
	defaultLimit = 50
	maxLimit     = 2000
)

// ReleasesHandler handles GET /api/releases requests.
type ReleasesHandler struct {
	releaseRepo repo.ReleaseRepository
}

// NewReleasesHandler creates a new ReleasesHandler.
func NewReleasesHandler(releaseRepo repo.ReleaseRepository) *ReleasesHandler {
	return &ReleasesHandler{releaseRepo: releaseRepo}
}

// GetRecentReleases handles the HTTP request to list recent releases.
func (h *ReleasesHandler) GetRecentReleases(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	releases, err := h.releaseRepo.GetRecent(r.Context(), limit, offset)
	if err != nil {
		slog.Error("failed to get recent releases", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"releases": releases,
		"limit":    limit,
		"offset":   offset,
	})
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = defaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= maxLimit {
			limit = parsed
		}
	}

	offset = 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
