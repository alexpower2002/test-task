package create

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type request struct {
	Status      string `json:"status"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AssigneeId  int    `json:"assignee_id"`
	TeamId      int    `json:"team_id"`
}

type response struct {
	Id          int       `json:"id"`
	Status      string    `json:"status"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatorId   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	AssigneeId  int       `json:"assignee_id"`
	TeamId      int       `json:"team_id"`
}

type Executor interface {
	Execute(ctx context.Context, in CreateInput) (*CreateResult, error)
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

	result, err := h.exec.Execute(r.Context(), CreateInput{
		CreatorId:   userId,
		Status:      req.Status,
		Title:       req.Title,
		Description: req.Description,
		AssigneeId:  req.AssigneeId,
		TeamId:      req.TeamId,
	})

	if err != nil {
		http.Error(w, "execution error", http.StatusBadRequest)

		return
	}

	resp := response{
		Id:          result.Id,
		Status:      result.Status,
		Title:       result.Title,
		Description: result.Description,
		CreatorId:   result.CreatorId,
		CreatedAt:   result.CreatedAt,
		AssigneeId:  result.AssigneeId,
		TeamId:      result.TeamId,
	}

	respBody, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}
