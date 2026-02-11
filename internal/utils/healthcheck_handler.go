package utils

import (
	"net/http"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func NewHealthcheckHandler() http.Handler {
	return http.HandlerFunc(HealthCheckHandler)
}
