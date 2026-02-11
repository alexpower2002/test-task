package list

import (
	"context"

	"mkk-luna-test-task/internal/team"
)

type ListResult struct {
	Teams []TeamItem
}

type TeamItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type teamLister interface {
	ListTeams(ctx context.Context, userId int) ([]*team.Model, error)
}

type executor struct {
	teamLister teamLister
}

func NewExecutor(teamLister teamLister) *executor {
	return &executor{teamLister: teamLister}
}

func (e *executor) Execute(ctx context.Context, userId int) (*ListResult, error) {
	teams, err := e.teamLister.ListTeams(ctx, userId)

	if err != nil {
		return nil, err
	}

	items := make([]TeamItem, 0, len(teams))

	for _, t := range teams {
		items = append(items, TeamItem{Id: t.Id, Name: t.Name})
	}

	return &ListResult{Teams: items}, nil
}
