package list

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type response struct {
	History         []HistoryItem `json:"history"`
	NextStartFromId int           `json:"next_start_from_id,omitempty"`
	HasMore         bool          `json:"has_more"`
}

type Executor interface {
	Execute(ctx context.Context, in ListInput) (*ListResult, error)
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

	taskIdStr := chi.URLParam(r, "id")

	taskId, err := strconv.Atoi(taskIdStr)

	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)

		return
	}

	query := r.URL.Query()

	startFromId := 0

	if s := query.Get("start_from_id"); s != "" {
		startFromId, err = strconv.Atoi(s)

		if err != nil {
			http.Error(w, "invalid start_from_id", http.StatusBadRequest)

			return
		}
	}

	result, err := h.exec.Execute(r.Context(), ListInput{
		UserId:      userId,
		TaskId:      taskId,
		StartFromId: startFromId,
	})

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	resp := response{
		History:         result.History,
		NextStartFromId: result.NextStartFromId,
		HasMore:         result.HasMore,
	}

	respBody, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
