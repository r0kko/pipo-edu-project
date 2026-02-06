package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	HTTPRequests *prometheus.CounterVec
	HTTPLatency  *prometheus.HistogramVec
	Passes       *prometheus.CounterVec
	Users        *prometheus.CounterVec
	Guest        *prometheus.CounterVec
	registry     *prometheus.Registry
}

func New(registry *prometheus.Registry) *Metrics {
	m := &Metrics{
		HTTPRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		Passes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "passes_events_total",
				Help: "Pass events total",
			},
			[]string{"action"},
		),
		Users: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "users_events_total",
				Help: "User events total",
			},
			[]string{"action"},
		),
		Guest: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "guest_requests_events_total",
				Help: "Guest request events total",
			},
			[]string{"action"},
		),
		registry: registry,
	}

	registry.MustRegister(m.HTTPRequests, m.HTTPLatency, m.Passes, m.Users, m.Guest)
	return m
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		status := strconv.Itoa(rw.status)
		path := r.URL.Path
		m.HTTPRequests.WithLabelValues(r.Method, path, status).Inc()
		m.HTTPLatency.WithLabelValues(r.Method, path, status).Observe(time.Since(start).Seconds())
	})
}

func (m *Metrics) Handler() http.Handler {
	if m.registry == nil {
		return promhttp.Handler()
	}
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
