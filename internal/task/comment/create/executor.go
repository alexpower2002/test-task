package create

import (
	"context"
	"fmt"
	"time"

	"mkk-luna-test-task/internal/task/comment"
)

type taskCommentCreator interface {
	CreateTaskComment(ctx context.Context, commenterId int, taskId int, text string, createdAt time.Time) (*comment.Model, error)
}

type teamMembershipChecker interface {
	IsPartOfTeamByTaskId(userId, taskId int) (bool, error)
}

type executor struct {
	taskCommentCreator    taskCommentCreator
	teamMembershipChecker teamMembershipChecker
}

func NewExecutor(
	taskCommentCreator taskCommentCreator,
	teamMembershipChecker teamMembershipChecker,
) *executor {
	return &executor{
		taskCommentCreator:    taskCommentCreator,
		teamMembershipChecker: teamMembershipChecker,
	}
}

type CreateInput struct {
	CommenterId int
	TaskId      int
	Text        string
}

func (e *executor) Execute(ctx context.Context, in CreateInput) (id int, err error) {
	isMember, err := e.teamMembershipChecker.IsPartOfTeamByTaskId(in.CommenterId, in.TaskId)

	if err != nil {
		return 0, err
	}

	if !isMember {
		return 0, fmt.Errorf("forbidden")
	}

	model, err := e.taskCommentCreator.CreateTaskComment(ctx, in.CommenterId, in.TaskId, in.Text, time.Now())

	if err != nil {
		return 0, err
	}

	return model.Id, nil
}
