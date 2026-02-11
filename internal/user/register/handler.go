package register

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type request struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type response struct {
	UserId int `json:"user_id"`
}

type RegisterExecutor interface {
	Execute(ctx context.Context, in RegisterInput) (*RegisterResult, error)
}

type handler struct {
	exec RegisterExecutor
}

func NewHandler(exec RegisterExecutor) *handler {
	return &handler{exec: exec}
}

func (h *handler) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)

		return
	}

	var req request

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)

		return
	}

	result, err := h.exec.Execute(r.Context(), RegisterInput{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	resp := response{UserId: result.UserId}

	respBytes, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "could not marshal response", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}
