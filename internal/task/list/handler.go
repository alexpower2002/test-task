package list

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type response struct {
	Tasks           []TaskItem `json:"tasks"`
	NextStartFromId int        `json:"next_start_from_id,omitempty"`
	HasMore         bool       `json:"has_more"`
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

	query := r.URL.Query()

	teamIdStr := query.Get("team_id")

	if teamIdStr == "" {
		http.Error(w, "team_id is required", http.StatusBadRequest)

		return
	}

	teamId, err := strconv.Atoi(teamIdStr)

	if err != nil {
		http.Error(w, "invalid team_id", http.StatusBadRequest)

		return
	}

	status := query.Get("status")

	assigneeId := 0

	if s := query.Get("assignee_id"); s != "" {
		assigneeId, err = strconv.Atoi(s)

		if err != nil {
			http.Error(w, "invalid assignee_id", http.StatusBadRequest)

			return
		}
	}

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
		TeamId:      teamId,
		Status:      status,
		AssigneeId:  assigneeId,
		StartFromId: startFromId,
	})

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	resp := response{
		Tasks:           result.Tasks,
		NextStartFromId: result.NextStartFromId,
		HasMore:         result.HasMore,
	}

	respBody, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}
