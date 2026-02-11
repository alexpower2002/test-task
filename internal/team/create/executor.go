package create

import (
	"context"

	"mkk-luna-test-task/internal/team"
)

type teamCreatorAndOwnerMaker interface {
	CreateTeamAndMakeUserItsOwner(ctx context.Context, name string, userId int) (*team.Model, error)
}

type executor struct {
	teamCreatorAndOwnerMaker teamCreatorAndOwnerMaker
}

func NewExecutor(teamCreatorAndOwnerMaker teamCreatorAndOwnerMaker) *executor {
	return &executor{teamCreatorAndOwnerMaker: teamCreatorAndOwnerMaker}
}

type CreateInput struct {
	UserId int
	Name   string
}

type CreateResult struct {
	Id   int
	Name string
}

func (e *executor) Execute(ctx context.Context, in CreateInput) (*CreateResult, error) {
	model, err := e.teamCreatorAndOwnerMaker.CreateTeamAndMakeUserItsOwner(ctx, in.Name, in.UserId)

	if err != nil {
		return nil, err
	}

	return &CreateResult{Id: model.Id, Name: model.Name}, nil
}
