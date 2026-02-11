package list

import (
	"context"
	"fmt"

	"mkk-luna-test-task/internal/task"
)

const defaultLimit = 100

type ListResult struct {
	Tasks           []TaskItem
	NextStartFromId int
	HasMore         bool
}

type TaskItem struct {
	Id          int    `json:"id"`
	Status      string `json:"status"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatorId   int    `json:"creator_id"`
	AssigneeId  int    `json:"assignee_id"`
	TeamId      int    `json:"team_id"`
}

type taskLister interface {
	ListTasks(ctx context.Context, teamId int, status string, assigneeId int, startFromId int, limit int) ([]*task.Model, error)
}

type teamMembershipChecker interface {
	CheckUserIsInTeam(userId, teamId int) (bool, error)
}

type taskCacheLister interface {
	ListTasksFromCache(ctx context.Context, teamId int, limit int, startFromID *int) (tasks []task.Model, hit bool, err error)
}

type taskCacheWriter interface {
	CacheTasks(ctx context.Context, teamId int, tasks []task.Model) error
}

type executor struct {
	taskLister            taskLister
	teamMembershipChecker teamMembershipChecker
	cacheLister           taskCacheLister
	cacheWriter           taskCacheWriter
}

func NewExecutor(
	taskLister taskLister,
	teamMembershipChecker teamMembershipChecker,
	cacheLister taskCacheLister,
	cacheWriter taskCacheWriter,
) *executor {
	return &executor{
		taskLister:            taskLister,
		teamMembershipChecker: teamMembershipChecker,
		cacheLister:           cacheLister,
		cacheWriter:           cacheWriter,
	}
}

type ListInput struct {
	UserId      int
	TeamId      int
	Status      string
	AssigneeId  int
	StartFromId int
}

func (e *executor) Execute(ctx context.Context, in ListInput) (*ListResult, error) {
	isMember, err := e.teamMembershipChecker.CheckUserIsInTeam(in.UserId, in.TeamId)

	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, fmt.Errorf("forbidden")
	}

	listFromCache := (in.Status == "" && in.AssigneeId == 0)

	var tasks []*task.Model
	hasMore := false
	nextStartFromId := 0

	// Предполагаю, что самое нагруженное это не поиск, поэтому кэш только простого отображения.
	if listFromCache && e.cacheLister != nil {
		var startFromID *int = nil

		if in.StartFromId > 0 {
			startFromID = &in.StartFromId
		}

		cachedTasks, hit, err := e.cacheLister.ListTasksFromCache(ctx, in.TeamId, defaultLimit+1, startFromID)

		if err != nil {
			return nil, err
		}

		if hit && cachedTasks != nil {
			results := cachedTasks

			if len(cachedTasks) > defaultLimit {
				hasMore = true
				nextStartFromId = cachedTasks[defaultLimit].Id
				results = cachedTasks[:defaultLimit]
			}

			items := make([]TaskItem, 0, len(results))

			for _, t := range results {
				items = append(items, TaskItem{
					Id:          t.Id,
					Status:      t.Status,
					Title:       t.Title,
					Description: t.Description,
					CreatorId:   t.CreatorId,
					AssigneeId:  t.AssigneeId,
					TeamId:      t.TeamId,
				})
			}

			return &ListResult{
				Tasks:           items,
				NextStartFromId: nextStartFromId,
				HasMore:         hasMore,
			}, nil
		}
	}

	dbTasks, err := e.taskLister.ListTasks(ctx, in.TeamId, in.Status, in.AssigneeId, in.StartFromId, defaultLimit+1)

	if err != nil {
		return nil, err
	}

	tasks = dbTasks

	if listFromCache {
		modelTasks := make([]task.Model, 0, len(tasks))

		loopLen := len(tasks)

		if loopLen > defaultLimit {
			loopLen = defaultLimit
		}

		for _, t := range tasks[:loopLen] {
			if t != nil {
				modelTasks = append(modelTasks, *t)
			}
		}

		_ = e.cacheWriter.CacheTasks(ctx, in.TeamId, modelTasks)
	}

	results := tasks

	if len(tasks) > defaultLimit {
		hasMore = true
		nextStartFromId = tasks[defaultLimit].Id
		results = tasks[:defaultLimit]
	}

	items := make([]TaskItem, 0, len(results))

	for _, t := range results {
		items = append(items, TaskItem{
			Id:          t.Id,
			Status:      t.Status,
			Title:       t.Title,
			Description: t.Description,
			CreatorId:   t.CreatorId,
			AssigneeId:  t.AssigneeId,
			TeamId:      t.TeamId,
		})
	}

	return &ListResult{
		Tasks:           items,
		NextStartFromId: nextStartFromId,
		HasMore:         hasMore,
	}, nil
}
