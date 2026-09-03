package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	_metricsNamespace = "http"
	_metricsSubsystem = ""
)

// Metrics collects Prometheus metrics for the HTTP transport and exposes them.
type Metrics struct {
	requests   *prometheus.CounterVec
	duration   *prometheus.HistogramVec
	inProgress *prometheus.GaugeVec
}

// NewMetrics registers the HTTP collectors for the given service on the default registry.
func NewMetrics(serviceName string) *Metrics {
	constLabels := prometheus.Labels{"service": serviceName}

	return &Metrics{
		requests: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        prometheus.BuildFQName(_metricsNamespace, _metricsSubsystem, "requests_total"),
			Help:        "Count all http requests by status code, method and path.",
			ConstLabels: constLabels,
		}, []string{"status", "method", "path"}),
		duration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:        prometheus.BuildFQName(_metricsNamespace, _metricsSubsystem, "request_duration_seconds"),
			Help:        "Duration of all http requests by status code, method and path.",
			ConstLabels: constLabels,
			Buckets:     prometheus.DefBuckets,
		}, []string{"status", "method", "path"}),
		inProgress: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name:        prometheus.BuildFQName(_metricsNamespace, _metricsSubsystem, "requests_in_progress_total"),
			Help:        "All the requests in progress.",
			ConstLabels: constLabels,
		}, []string{"method"}),
	}
}

// Middleware instruments every request handled after it is registered.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		start := time.Now()

		m.inProgress.WithLabelValues(method).Inc()

		// Deferred so a panicking handler still releases the gauge and records the request.
		defer func() {
			m.inProgress.WithLabelValues(method).Dec()

			path := ctx.FullPath()
			if path == "" {
				path = ctx.Request.URL.Path
			}

			status := strconv.Itoa(ctx.Writer.Status())

			m.requests.WithLabelValues(status, method, path).Inc()
			m.duration.WithLabelValues(status, method, path).Observe(time.Since(start).Seconds())
		}()

		ctx.Next()
	}
}

// Handler exposes the collected metrics in the Prometheus text format.
func (m *Metrics) Handler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
