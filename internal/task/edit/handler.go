package edit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type request struct {
	Status      string `json:"status"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AssigneeId  int    `json:"assignee_id"`
}

type response struct {
	Success bool `json:"success"`
}

type Executor interface {
	Execute(ctx context.Context, in EditInput) (bool, error)
}

type handler struct {
	exec Executor
}

func NewHandler(exec Executor) *handler {
	return &handler{
		exec: exec,
	}
}

func (h *handler) Handle(w http.ResponseWriter, r *http.Request) {
	userIdAny := r.Context().Value("userId")

	userId, ok := userIdAny.(int)

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	idStr := chi.URLParam(r, "id")

	taskId, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)

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
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	ok, err = h.exec.Execute(r.Context(), EditInput{
		UserId:      userId,
		TaskId:      taskId,
		Status:      req.Status,
		Title:       req.Title,
		Description: req.Description,
		AssigneeId:  req.AssigneeId,
	})

	if err != nil || !ok {
		resp := response{Success: false}

		respBody, marshalErr := json.Marshal(resp)

		if marshalErr != nil {
			http.Error(w, "error", http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(respBody)

		return
	}

	resp := response{Success: true}
	respBody, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
