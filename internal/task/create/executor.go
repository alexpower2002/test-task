package create

import (
	"context"
	"fmt"
	"time"

	"mkk-luna-test-task/internal/task"
)

type taskCreator interface {
	CreateTask(ctx context.Context, status, title, description string, creatorId, assigneeId, teamId int, createdAt time.Time) (*task.Model, error)
}

type teamMembershipChecker interface {
	CheckUserIsInTeam(userId, teamId int) (bool, error)
}

type executor struct {
	taskCreator           taskCreator
	teamMembershipChecker teamMembershipChecker
}

func NewExecutor(taskCreator taskCreator, teamMembershipChecker teamMembershipChecker) *executor {
	return &executor{
		taskCreator:           taskCreator,
		teamMembershipChecker: teamMembershipChecker,
	}
}

type CreateInput struct {
	CreatorId   int
	Status      string
	Title       string
	Description string
	AssigneeId  int
	TeamId      int
}

type CreateResult struct {
	Id          int       `json:"id"`
	Status      string    `json:"status"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatorId   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	AssigneeId  int       `json:"assignee_id"`
	TeamId      int       `json:"team_id"`
}

func (e *executor) Execute(ctx context.Context, in CreateInput) (*CreateResult, error) {
	isMember, err := e.teamMembershipChecker.CheckUserIsInTeam(in.CreatorId, in.TeamId)

	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, fmt.Errorf("forbidden")
	}

	now := time.Now()

	model, err := e.taskCreator.CreateTask(ctx, in.Status, in.Title, in.Description, in.CreatorId, in.AssigneeId, in.TeamId, now)

	if err != nil {
		return nil, err
	}

	return &CreateResult{
		Id:          model.Id,
		Status:      model.Status,
		Title:       model.Title,
		Description: model.Description,
		CreatorId:   model.CreatorId,
		CreatedAt:   model.CreatedAt,
		AssigneeId:  model.AssigneeId,
		TeamId:      model.TeamId,
	}, nil
}
