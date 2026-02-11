package utils

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	startupCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_startups_total",
			Help: "Total number of times the application has started.",
		},
	)

	httpRequestsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests received.",
		},
		[]string{"method", "route"},
	)

	httpErrorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors encountered.",
		},
		[]string{"method", "route", "status"},
	)

	httpResponseTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_duration_ms",
			Help:    "Histogram of response times for HTTP requests.",
			Buckets: prometheus.LinearBuckets(100, 100, 60),
		},
		[]string{"method", "route"},
	)
)

type metrics struct{}

func InitMetrics() (metrics, error) {
	if err := prometheus.Register(startupCounter); err != nil {
		return metrics{}, err
	}

	if err := prometheus.Register(httpRequestsCounter); err != nil {
		return metrics{}, err
	}

	if err := prometheus.Register(httpErrorsCounter); err != nil {
		return metrics{}, err
	}

	if err := prometheus.Register(httpResponseTimeHistogram); err != nil {
		return metrics{}, err
	}

	return metrics{}, nil
}

func (metrics) IncStartup() {
	startupCounter.Inc()
}

func (metrics) IncHTTPRequest(method, route string) {
	httpRequestsCounter.WithLabelValues(method, route).Inc()
}

func (metrics) IncHTTPError(method, route string, statusCode int) {
	httpErrorsCounter.WithLabelValues(method, route, strconv.Itoa(statusCode)).Inc()
}

func (metrics) ObserveHTTPResponseTime(method, route string, duration time.Duration) {
	httpResponseTimeHistogram.WithLabelValues(method, route).Observe(float64(duration.Milliseconds()))
}
