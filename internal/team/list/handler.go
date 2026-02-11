package list

import (
	"context"
	"encoding/json"
	"net/http"
)

type response struct {
	Teams []TeamItem `json:"teams"`
}

type Executor interface {
	Execute(ctx context.Context, userId int) (*ListResult, error)
}

type Handler struct {
	exec Executor
}

func NewHandler(exec Executor) *Handler {
	return &Handler{
		exec: exec,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userIdAny := r.Context().Value("userId")

	userId, ok := userIdAny.(int)

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	result, err := h.exec.Execute(r.Context(), userId)

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	resp := response{Teams: result.Teams}

	respBody, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
