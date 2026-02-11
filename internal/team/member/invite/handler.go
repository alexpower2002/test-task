package invite

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"mkk-luna-test-task/internal/team/member"
)

type request struct {
	UserId int         `json:"user_id"`
	Role   member.Role `json:"role"`
}

type response struct {
	UserId int         `json:"user_id"`
	TeamId int         `json:"team_id"`
	Role   member.Role `json:"role"`
}

type handler struct {
	exec *executor
}

func NewHandler(exec *executor) *handler {
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

	teamIdStr := chi.URLParam(r, "id")

	teamId, err := strconv.Atoi(teamIdStr)

	if err != nil {
		http.Error(w, "invalid team id", http.StatusBadRequest)

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

	result, err := h.exec.Execute(r.Context(), InviteInput{
		InviterUserId: userId,
		TeamId:        teamId,
		UserId:        req.UserId,
		Role:          req.Role,
	})

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)

		return
	}

	resp := response{UserId: result.UserId, TeamId: result.TeamId, Role: result.Role}

	respBody, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "json error", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(respBody)
}
