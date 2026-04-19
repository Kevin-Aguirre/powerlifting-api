package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by method, route, and status code.",
		},
		[]string{"method", "route", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency distribution.",
			Buckets: []float64{.001, .005, .01, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"method", "route"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

// metricsMiddleware records per-route request counts and latencies.
// Uses the matched chi route pattern (e.g. /lifters/{lifterName}) to keep cardinality bounded.
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(sw, r)
		dur := time.Since(start).Seconds()

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = r.URL.Path
		}
		httpRequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(sw.status)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, route).Observe(dur)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func prometheusHandler() http.Handler {
	return promhttp.Handler()
}
