package utils

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

type CircuitBreakerRoundTripper struct {
	failures         int
	failureThreshold int
	state            CircuitState
	lastFailure      time.Time
	resetTimeout     time.Duration
	mu               sync.Mutex
}

func NewCircuitBreakerRoundTripper(failureThreshold int, resetTimeout time.Duration) *CircuitBreakerRoundTripper {
	return &CircuitBreakerRoundTripper{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            Closed,
	}
}

func (cb *CircuitBreakerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	cb.mu.Lock()

	switch cb.state {
	case Open:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = HalfOpen
		} else {
			cb.mu.Unlock()

			return nil, fmt.Errorf("circuit breaker is open")
		}
	}

	cb.mu.Unlock()

	resp, err := http.DefaultTransport.RoundTrip(req)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil || (resp != nil && resp.StatusCode >= 400) {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.state == HalfOpen || cb.failures >= cb.failureThreshold {
			cb.state = Open
		}
	} else {
		if cb.state == HalfOpen {
			cb.state = Closed
			cb.failures = 0
		} else if cb.state == Closed {
			cb.failures = 0
		}
	}

	return resp, err
}

type CircuitBreakerHTTPClient struct {
	client *http.Client
}

func NewCircuitBreakerHTTPClient(failureThreshold int, resetTimeout time.Duration) *CircuitBreakerHTTPClient {
	rt := NewCircuitBreakerRoundTripper(failureThreshold, resetTimeout)

	return &CircuitBreakerHTTPClient{
		client: &http.Client{Transport: rt},
	}
}

func (cb *CircuitBreakerHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return cb.client.Do(req)
}
