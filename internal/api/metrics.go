package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Application metrics, exposed on GET /metrics for Prometheus to scrape.
var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "miniclaude_http_requests_total",
		Help: "Total HTTP requests, by method, route and status code.",
	}, []string{"method", "route", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "miniclaude_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds, by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})

	authEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "miniclaude_auth_events_total",
		Help: "Authentication events, by type (register/login) and outcome (success/failure).",
	}, []string{"event", "outcome"})

	syncEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "miniclaude_sync_events_total",
		Help: "Data sync events, by type (export/import).",
	}, []string{"event"})
)

// metricsMiddleware records request count and latency for every request.
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		httpRequests.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		httpDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}

// metricsHandler serves the Prometheus exposition endpoint.
func metricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
