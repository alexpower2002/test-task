package create

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type request struct {
	Name string `json:"name"`
}

type response struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Executor interface {
	Execute(ctx context.Context, in CreateInput) (*CreateResult, error)
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

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)

		return
	}

	var req request

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "json error", http.StatusBadRequest)

		return
	}

	result, err := h.exec.Execute(r.Context(), CreateInput{UserId: userId, Name: req.Name})

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	resp := response{Id: result.Id, Name: result.Name}

	respBody, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}
