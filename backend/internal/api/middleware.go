package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// originAllowed reports whether origin matches one of the allowed patterns.
// A pattern is an exact origin ("https://example.com") or has a leading "*"
// that stands for part of the first DNS label (letters, digits, hyphens; never
// a dot), followed by either:
//   - ".": "https://*.example.com" matches "https://abc123.example.com" but not
//     "https://example.com" or "https://a.b.example.com".
//   - "-": "https://*-app.acct.workers.dev" matches Cloudflare Workers preview
//     URLs such as "https://my-branch-app.acct.workers.dev", but not other
//     Workers on the same account subdomain.
func originAllowed(patterns []string, origin string) bool {
	if origin == "" {
		return false
	}
	for _, p := range patterns {
		if p == origin {
			return true
		}
		scheme, suffix, ok := strings.Cut(p, "://*")
		if !ok || scheme == "" || len(suffix) < 2 || (suffix[0] != '.' && suffix[0] != '-') {
			continue
		}
		prefix := scheme + "://"
		if len(origin) <= len(prefix)+len(suffix) ||
			!strings.HasPrefix(origin, prefix) || !strings.HasSuffix(origin, suffix) {
			continue
		}
		if isDNSLabel(origin[len(prefix) : len(origin)-len(suffix)]) {
			return true
		}
	}
	return false
}

// isDNSLabel reports whether s is a single lowercase DNS label (no dots).
func isDNSLabel(s string) bool {
	if len(s) == 0 || len(s) > 63 || s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	for _, c := range s {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return true
}

// corsMiddleware adds CORS headers to allow requests from the frontend.
func corsMiddleware(next http.Handler) http.Handler {
	// ALLOWED_ORIGINS is a comma-separated list of exact origins and/or wildcard
	// patterns (see originAllowed), e.g.
	// "https://gopackage-chronicle.anphse.workers.dev,https://*-gopackage-chronicle.anphse.workers.dev".
	// In development, it defaults to allowing localhost.
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		// Development default: allow common local dev ports
		allowedOriginsEnv = "http://localhost:5173,http://localhost:3000,http://localhost:4173"
	}

	allowedOrigins := strings.Split(allowedOriginsEnv, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Set CORS headers if origin is allowed
		if originAllowed(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if allowedOriginsEnv == "*" {
			// Only allow wildcard if explicitly set in production
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// rateLimiter manages per-IP rate limiting using a token bucket algorithm.
type rateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// newRateLimiter creates a new rate limiter.
// rate: requests per second per IP
// burst: maximum burst size
func newRateLimiter(r rate.Limit, b int) *rateLimiter {
	return &rateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter returns the rate limiter for a given IP address.
func (rl *rateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// rateLimitMiddleware limits requests per IP address.
func rateLimitMiddleware(rl *rateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP (handle X-Forwarded-For for proxies)
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.Header.Get("X-Real-IP")
			}
			if ip == "" {
				ip = r.RemoteAddr
			}

			limiter := rl.getLimiter(ip)
			if !limiter.Allow() {
				slog.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path)
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// timeoutMiddleware adds a timeout to all requests.
func timeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Request completed successfully
			case <-ctx.Done():
				// Request timed out
				slog.Error("request timeout", "path", r.URL.Path, "timeout", timeout)
				http.Error(w, "request timeout", http.StatusGatewayTimeout)
			}
		})
	}
}

// loggingMiddleware logs all incoming requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		slog.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
