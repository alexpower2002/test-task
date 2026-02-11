package login

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
	Token string `json:"token"`
}

type LoginExecutor interface {
	Execute(ctx context.Context, in LoginInput) (*LoginResult, error)
}

type handler struct {
	exec LoginExecutor
}

func NewHandler(exec LoginExecutor) *handler {
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

	result, err := h.exec.Execute(r.Context(), LoginInput{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	resp := response{Token: result.Token}

	respBytes, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "could not marshal response", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respBytes)
}
