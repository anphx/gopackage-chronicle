package api

import (
	"database/sql"
	"net/http"

	"github.com/anphx/gopackage-chronicles/internal/api/handlers"
	repo "github.com/anphx/gopackage-chronicles/internal/repository"
)

// NewServer creates a new HTTP server with all routes configured.
func NewServer(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	// Initialize repositories
	packageRepo := repo.NewPackageRepository(db)
	releaseRepo := repo.NewReleaseRepository(db)

	// Initialize handlers
	releasesHandler := handlers.NewReleasesHandler(releaseRepo)
	packagesHandler := handlers.NewPackagesHandler(packageRepo, releaseRepo)

	// Register routes
	// Releases
	mux.HandleFunc("GET /api/releases", releasesHandler.GetRecentReleases)

	// Packages
	mux.HandleFunc("GET /api/packages", packagesHandler.HandleList)
	mux.HandleFunc("GET /api/packages/{name...}", packagesHandler.HandleDetail)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return mux
}
