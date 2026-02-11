package list

import (
	"context"
	"fmt"
	"time"

	"mkk-luna-test-task/internal/task/history"
)

const defaultLimit = 100

type ListResult struct {
	History         []HistoryItem
	NextStartFromId int
	HasMore         bool
}

type HistoryItem struct {
	Id          int       `json:"id"`
	TaskId      int       `json:"task_id"`
	Status      string    `json:"status"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatorId   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	AssigneeId  int       `json:"assignee_id"`
	TeamId      int       `json:"team_id"`
	ChangedBy   int       `json:"changed_by"`
	ChangedAt   time.Time `json:"changed_at"`
}

type historyLister interface {
	ListHistory(ctx context.Context, taskId int, startFromId int, limit int) ([]*history.Model, error)
}

type taskMembershipChecker interface {
	IsPartOfTeamByTaskId(userId, taskId int) (bool, error)
}

type executor struct {
	historyLister         historyLister
	taskMembershipChecker taskMembershipChecker
}

func NewExecutor(historyLister historyLister, taskMembershipChecker taskMembershipChecker) *executor {
	return &executor{
		historyLister:         historyLister,
		taskMembershipChecker: taskMembershipChecker,
	}
}

type ListInput struct {
	UserId      int
	TaskId      int
	StartFromId int
}

func (e *executor) Execute(ctx context.Context, in ListInput) (*ListResult, error) {
	isMember, err := e.taskMembershipChecker.IsPartOfTeamByTaskId(in.UserId, in.TaskId)

	if err != nil {
		return nil, fmt.Errorf("failed to check team membership: %w", err)
	}

	if !isMember {
		return nil, fmt.Errorf("forbidden")
	}

	histories, err := e.historyLister.ListHistory(ctx, in.TaskId, in.StartFromId, defaultLimit+1)

	if err != nil {
		return nil, err
	}

	hasMore := false
	nextStartFromId := 0
	results := histories

	if len(histories) > defaultLimit {
		hasMore = true
		nextStartFromId = histories[defaultLimit].Id
		results = histories[:defaultLimit]
	}

	items := make([]HistoryItem, 0, len(results))

	for _, h := range results {
		items = append(items, HistoryItem{
			Id:          h.Id,
			TaskId:      h.TaskId,
			Status:      h.Status,
			Title:       h.Title,
			Description: h.Description,
			CreatorId:   h.CreatorId,
			CreatedAt:   h.CreatedAt,
			AssigneeId:  h.AssigneeId,
			TeamId:      h.TeamId,
			ChangedBy:   h.ChangedBy,
			ChangedAt:   h.ChangedAt,
		})
	}

	return &ListResult{
		History:         items,
		NextStartFromId: nextStartFromId,
		HasMore:         hasMore,
	}, nil
}
