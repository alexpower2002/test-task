package utils

import (
	"net/http"
	"time"
)

type metricsMiddlewareHandler struct {
	metrics metrics
	next    http.Handler
}

func NewMetricsMiddlewareHandler(metrics metrics, next http.Handler) *metricsMiddlewareHandler {
	return &metricsMiddlewareHandler{
		metrics: metrics,
		next:    next,
	}
}

func (m *metricsMiddlewareHandler) Handle(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// нужно поймать запись статус-кода.
	rr := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

	route := r.URL.Path

	m.metrics.IncHTTPRequest(r.Method, route)

	m.next.ServeHTTP(rr, r)

	duration := time.Since(startTime)

	m.metrics.ObserveHTTPResponseTime(r.Method, route, duration)

	if rr.statusCode >= 400 {
		m.metrics.IncHTTPError(r.Method, route, rr.statusCode)
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if rr.statusCode == 0 {
		rr.statusCode = http.StatusOK
	}

	return rr.ResponseWriter.Write(b)
}
