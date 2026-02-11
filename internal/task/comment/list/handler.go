package list

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type response struct {
	Comments        []CommentItem `json:"comments"`
	NextStartFromId int           `json:"next_start_from_id,omitempty"`
	HasMore         bool          `json:"has_more"`
}

type Executor interface {
	Execute(ctx context.Context, taskId, startFromId int) (*ListResult, error)
}

func NewHandler(exec Executor) *handler {
	return &handler{exec: exec}
}

type handler struct {
	exec Executor
}

func (h *handler) Handle(w http.ResponseWriter, r *http.Request) {
	taskIdStr := chi.URLParam(r, "id")

	taskId, err := strconv.Atoi(taskIdStr)

	if err != nil {
		http.Error(w, "invalid task_id", http.StatusBadRequest)

		return
	}

	startFromId := 0

	if s := r.URL.Query().Get("start_from_id"); s != "" {
		startFromId, err = strconv.Atoi(s)

		if err != nil {
			http.Error(w, "invalid start_from_id", http.StatusBadRequest)

			return
		}
	}

	result, err := h.exec.Execute(r.Context(), taskId, startFromId)

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	resp := response{
		Comments:        result.Comments,
		NextStartFromId: result.NextStartFromId,
		HasMore:         result.HasMore,
	}

	body, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
