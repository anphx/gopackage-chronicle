package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	repo "github.com/anphx/gopackage-chronicles/internal/repository"
)

// PackagesHandler handles package-related requests.
type PackagesHandler struct {
	packageRepo repo.PackageRepository
	releaseRepo repo.ReleaseRepository
}

// NewPackagesHandler creates a new PackagesHandler.
func NewPackagesHandler(packageRepo repo.PackageRepository, releaseRepo repo.ReleaseRepository) *PackagesHandler {
	return &PackagesHandler{
		packageRepo: packageRepo,
		releaseRepo: releaseRepo,
	}
}

// HandleList handles GET /api/packages - lists all packages.
func (h *PackagesHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	packages, err := h.packageRepo.List(r.Context(), limit, offset)
	if err != nil {
		slog.Error("failed to list packages", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"packages": packages,
		"limit":    limit,
		"offset":   offset,
	})
}

// HandleDetail handles GET /api/packages/{name...} - gets package details and releases.
func (h *PackagesHandler) HandleDetail(w http.ResponseWriter, r *http.Request) {
	packageName := r.PathValue("name")
	if packageName == "" {
		http.Error(w, "package name is required", http.StatusBadRequest)
		return
	}

	pkg, err := h.packageRepo.GetByPath(r.Context(), packageName)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			http.Error(w, "package not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to get package", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	limit, offset := parsePagination(r)

	releases, err := h.releaseRepo.GetByPackageID(r.Context(), pkg.ID, limit, offset)
	if err != nil {
		slog.Error("failed to get package releases", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"package":  pkg,
		"releases": releases,
		"limit":    limit,
		"offset":   offset,
	})
}
