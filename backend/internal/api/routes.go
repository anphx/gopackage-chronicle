package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/anphx/gopackage-chronicles/internal/api/handlers"
	repo "github.com/anphx/gopackage-chronicles/internal/repository"
	"golang.org/x/time/rate"
)

// NewServer creates a new HTTP server with all routes configured.
func NewServer(db *sql.DB) http.Handler {
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

	// Apply rate limiter: 10 requests per second per IP, burst of 20.
	// To prevent abuse of this publicly accessible API, we aim to cap the number of requests originating from each IP address.
	limiter := newRateLimiter(rate.Limit(10), 20)

	// Apply middleware chains.
	var handler http.Handler = mux
	handler = timeoutMiddleware(30 * time.Second)(handler) // Request timeout (innermost)
	handler = rateLimitMiddleware(limiter)(handler)        // Rate limiting
	handler = loggingMiddleware(handler)                   // Request logging
	handler = corsMiddleware(handler)                      // CORS headers (outermost)

	return handler
}
