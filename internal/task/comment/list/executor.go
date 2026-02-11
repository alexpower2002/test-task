package list

import (
	"context"
	"time"

	"mkk-luna-test-task/internal/task/comment"
)

const defaultLimit = 100

type ListResult struct {
	Comments        []CommentItem
	NextStartFromId int
	HasMore         bool
}

type CommentItem struct {
	Id          int       `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	CommenterId int       `json:"commenter_id"`
	TaskId      int       `json:"task_id"`
	Text        string    `json:"text"`
}

type taskCommentLister interface {
	ListTaskComments(ctx context.Context, taskId int, startFromId int, limit int) ([]comment.Model, error)
}

type executor struct {
	taskCommentLister taskCommentLister
}

func NewExecutor(taskCommentLister taskCommentLister) *executor {
	return &executor{taskCommentLister: taskCommentLister}
}

func (e *executor) Execute(ctx context.Context, taskId, startFromId int) (*ListResult, error) {
	comments, err := e.taskCommentLister.ListTaskComments(ctx, taskId, startFromId, defaultLimit+1)

	if err != nil {
		return nil, err
	}

	hasMore := false
	nextStartFromId := 0
	results := comments

	if len(comments) > defaultLimit {
		hasMore = true
		nextStartFromId = comments[defaultLimit].Id
		results = comments[:defaultLimit]
	}

	items := make([]CommentItem, 0, len(results))

	for _, c := range results {
		items = append(items, CommentItem{
			Id:          c.Id,
			CreatedAt:   c.CreatedAt,
			CommenterId: c.CommenterId,
			TaskId:      c.TaskId,
			Text:        c.Text,
		})
	}

	return &ListResult{
		Comments:        items,
		NextStartFromId: nextStartFromId,
		HasMore:         hasMore,
	}, nil
}
