package create

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type request struct {
	Text string `json:"text"`
}

type response struct {
	Id int `json:"id"`
}

type Executor interface {
	Execute(ctx context.Context, in CreateInput) (id int, err error)
}

type handler struct {
	exec Executor
}

func NewHandler(
	exec Executor,
) *handler {
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

	taskIdStr := chi.URLParam(r, "id")

	taskId, err := strconv.Atoi(taskIdStr)

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
		http.Error(w, "json error", http.StatusBadRequest)

		return
	}

	id, err := h.exec.Execute(r.Context(), CreateInput{
		CommenterId: userId,
		TaskId:      taskId,
		Text:        req.Text,
	})

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	resp := response{Id: id}

	respBody, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}
