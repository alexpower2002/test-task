package edit

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"mkk-luna-test-task/internal/task"
	"mkk-luna-test-task/internal/task/history"
)

type taskEditor interface {
	EditTask(ctx context.Context, id int, status, title, description string, assigneeId int) error
}

type taskGetter interface {
	GetTaskById(ctx context.Context, id int) (*task.Model, error)
}

type historySaver interface {
	CreateHistory(ctx context.Context, task *task.Model, changedBy int, changedAt time.Time) (*history.Model, error)
}

type taskMembershipChecker interface {
	IsPartOfTeamByTaskId(userId, taskId int) (bool, error)
}

type cacheUpdater interface {
	UpdateTaskInCache(ctx context.Context, t task.Model) error
}

type executor struct {
	taskEditor            taskEditor
	taskGetter            taskGetter
	historySaver          historySaver
	taskMembershipChecker taskMembershipChecker
	cacheUpdater          cacheUpdater
}

func NewExecutor(
	taskEditor taskEditor,
	taskGetter taskGetter,
	historySaver historySaver,
	taskMembershipChecker taskMembershipChecker,
	cacheUpdater cacheUpdater,
) *executor {
	return &executor{
		taskEditor:            taskEditor,
		taskGetter:            taskGetter,
		historySaver:          historySaver,
		taskMembershipChecker: taskMembershipChecker,
		cacheUpdater:          cacheUpdater,
	}
}

type EditInput struct {
	UserId      int
	TaskId      int
	Status      string
	Title       string
	Description string
	AssigneeId  int
}

func (e *executor) Execute(ctx context.Context, in EditInput) (bool, error) {
	oldTask, err := e.taskGetter.GetTaskById(ctx, in.TaskId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("task not found")
		}

		return false, err
	}

	isMember, err := e.taskMembershipChecker.IsPartOfTeamByTaskId(in.UserId, in.TaskId)

	if err != nil {
		return false, fmt.Errorf("failed to check team membership: %w", err)
	}

	if !isMember {
		return false, fmt.Errorf("forbidden")
	}

	_, err = e.historySaver.CreateHistory(ctx, oldTask, in.UserId, time.Now())

	if err != nil {
		return false, fmt.Errorf("failed to save task history: %w", err)
	}

	err = e.taskEditor.EditTask(ctx, in.TaskId, in.Status, in.Title, in.Description, in.AssigneeId)

	if err != nil {
		return false, err
	}

	updatedTask := *oldTask
	updatedTask.Status = in.Status
	updatedTask.Title = in.Title
	updatedTask.Description = in.Description
	updatedTask.AssigneeId = in.AssigneeId

	// В худшем случае (если не получилось) старая версия просто улетит из кэша по ttl.
	// Если нужны жёсткие гарантии консистентности, то нужно думать над архитектурой.
	_ = e.cacheUpdater.UpdateTaskInCache(ctx, updatedTask)

	return true, nil
}
