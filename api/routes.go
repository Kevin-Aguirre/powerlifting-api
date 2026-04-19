package api

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Kevin-Aguirre/powerlifting-api/data"
	handlers "github.com/Kevin-Aguirre/powerlifting-api/internal/clone"
)

func NewRouter(ds *data.DataStore) http.Handler {
	r := chi.NewRouter()

	r.Use(requestLogger)
	r.Use(metricsMiddleware)
	r.Use(rateLimitMiddleware(newRateLimiter(120, 20)))

	// health
	r.Get("/health", healthHandler(ds))

	// api index
	r.Get("/", apiIndex)

	// lifters
	r.Get("/lifters", handlers.GetLifters(ds))
	r.Get("/lifters/names", handlers.GetLifterNames(ds))
	r.Get("/lifters/search", handlers.SearchLifters(ds))
	r.Get("/lifters/top", handlers.GetTopLifters(ds))
	r.Get("/lifters/compare", handlers.CompareLifters(ds))
	r.Get("/lifters/{lifterName}", handlers.GetLifter(ds))
	r.Get("/lifters/{lifterName}/stats", handlers.GetLifterStats(ds))

	// meets
	r.Get("/meets", handlers.GetMeets(ds))
	r.Get("/meets/{federationName}", handlers.GetMeet(ds))
	r.Get("/meets/{federationName}/{meetName}/results", handlers.GetMeetResults(ds))

	// records
	r.Get("/records", handlers.GetRecords(ds))

	// federations
	r.Get("/federations", handlers.GetFederations(ds))

	// openapi spec
	r.Get("/openapi.json", openAPIHandler())

	// prometheus metrics
	r.Handle("/metrics", prometheusHandler())

	return r
}

// requestLogger logs each incoming request.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request", "method", r.Method, "path", r.URL.RequestURI(), "ip", realIP(r))
		next.ServeHTTP(w, r)
	})
}

// healthHandler returns the current status and last data update time.
func healthHandler(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		status := "ok"
		if db == nil {
			status = "loading"
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      status,
			"lastUpdated": ds.LastUpdated().UTC().Format(time.RFC3339),
		})
	}
}

// --- Rate limiter (per-IP token bucket) ---

type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*bucket
	rate    float64 // tokens per second
	burst   float64
}

type bucket struct {
	tokens   float64
	lastTime time.Time
}

// newRateLimiter creates a rate limiter allowing ratePerMin requests per minute
// per IP with a burst allowance.
func newRateLimiter(ratePerMin, burst int) *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string]*bucket),
		rate:    float64(ratePerMin) / 60.0,
		burst:   float64(burst),
	}
	go func() {
		for range time.Tick(5 * time.Minute) {
			rl.mu.Lock()
			for ip, b := range rl.clients {
				if time.Since(b.lastTime) > 5*time.Minute {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.clients[ip]
	if !ok {
		rl.clients[ip] = &bucket{tokens: rl.burst - 1, lastTime: time.Now()}
		return true
	}
	now := time.Now()
	b.tokens = min(rl.burst, b.tokens+now.Sub(b.lastTime).Seconds()*rl.rate)
	b.lastTime = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func rateLimitMiddleware(rl *rateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(realIP(r)) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// realIP extracts the client IP, respecting X-Forwarded-For headers.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.SplitN(xff, ",", 2)[0]
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func apiIndex(w http.ResponseWriter, r *http.Request) {
	index := map[string]interface{}{
		"name":    "Powerlifting API",
		"version": "1.0",
		"endpoints": map[string]string{
			"GET /health":                                                    "API health and data freshness",
			"GET /lifters?sex=M&equipment=Raw&sort=name&order=asc":           "List all lifters (paginated, filterable, sortable)",
			"GET /lifters/names?sex=M&equipment=Raw&sort=name&order=asc":     "List all lifter names (paginated, filterable, sortable)",
			"GET /lifters/search?q={query}&sort=name&order=asc":              "Search lifters by name (paginated, sortable)",
			"GET /lifters/top?sex=M&equipment=Raw&weightClass=83":            "Top lifters ranked by DOTS (paginated)",
			"GET /lifters/compare?a={nameA}&b={nameB}":                       "Compare two lifters side by side",
			"GET /lifters/{lifterName}":                                       "Get a single lifter by name",
			"GET /lifters/{lifterName}/stats":                                 "Career stats — meet count, federations, PR progression",
			"GET /meets?country=USA&from=2020-01-01&to=2023-12-31&sort=date": "List all meets (paginated, filterable, sortable)",
			"GET /meets/{federationName}?from=2020-01-01&sort=date":          "List meets by federation (paginated, filterable, sortable)",
			"GET /meets/{federationName}/{meetName}/results":                  "Get all entries for a specific meet (paginated)",
			"GET /records?sex=M&equipment=Raw&weightClass=83":                "All-time records by weight class (paginated)",
			"GET /federations?sort=name&order=asc":                           "List all federation names (paginated, sortable)",
			"GET /openapi.json":                                               "OpenAPI 3.0 specification",
		},
		"pagination": map[string]string{
			"limit":   "Number of results per page (default: 50, max: 200)",
			"offset":  "Number of results to skip (default: 0)",
			"example": "/lifters?limit=20&offset=100",
		},
		"unit_conversion": map[string]string{
			"parameter": "unit",
			"values":    "kg (default), lbs",
			"example":   "/lifters/John%20Doe?unit=lbs",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(index)
}
