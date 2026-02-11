package utils

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	rateLimitRequests = 100
	rateLimitDuration = 1 * time.Minute
)

type RateLimitingMiddleware struct {
	limiters sync.Map
	next     http.Handler
}

func NewRateLimitingMiddleware(next http.Handler) *RateLimitingMiddleware {
	return &RateLimitingMiddleware{
		next: next,
	}
}

func (m *RateLimitingMiddleware) getLimiter(key string) *rate.Limiter {
	limiter, ok := m.limiters.Load(key)

	if ok {
		return limiter.(*rate.Limiter)
	}

	limiter = rate.NewLimiter(rate.Every(rateLimitDuration/rateLimitRequests), rateLimitRequests)

	actual, _ := m.limiters.LoadOrStore(key, limiter)

	return actual.(*rate.Limiter)
}

func (m *RateLimitingMiddleware) Handle(w http.ResponseWriter, r *http.Request) {
	key := getUserIdOrIp(r)

	limiter := m.getLimiter(key)

	if !limiter.Allow() {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)

		return
	}

	m.next.ServeHTTP(w, r)
}

func getUserIdOrIp(r *http.Request) string {
	if userID := r.Context().Value("userId"); userID != nil {
		if v, ok := userID.(int); ok {
			return "user_id:" + strconv.Itoa(v)
		}
	}

	return "ip:" + getClientIP(r)
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")

	if xff != "" {
		parts := strings.Split(xff, ",")

		for _, p := range parts {
			ip := strings.TrimSpace(p)

			if ip != "" {
				return ip
			}
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
