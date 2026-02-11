package invite

import (
	"context"
	"fmt"

	"mkk-luna-test-task/internal/team/member"
)

type memberInviter interface {
	InviteMember(ctx context.Context, userId, teamId int, role member.Role) (*member.Model, error)
}

type memberInvitationChecker interface {
	CanUserInvite(ctx context.Context, teamId int, userId int) (bool, error)
}

type emailSender interface {
	SendEmail(address string, text string) error
}

type executor struct {
	memberInviter           memberInviter
	memberInvitationChecker memberInvitationChecker
	emailSender             emailSender
}

func NewExecutor(
	memberInviter memberInviter,
	memberInvitationChecker memberInvitationChecker,
	emailSender emailSender,
) *executor {
	return &executor{
		memberInviter:           memberInviter,
		memberInvitationChecker: memberInvitationChecker,
		emailSender:             emailSender,
	}
}

type InviteInput struct {
	InviterUserId int
	TeamId        int
	UserId        int
	Role          member.Role
	Email         string
}

type InviteResult struct {
	UserId int
	TeamId int
	Role   member.Role
}

const simpleEmailText = "You have been invited to new team!"

func (e *executor) Execute(ctx context.Context, in InviteInput) (*InviteResult, error) {
	canInvite, err := e.memberInvitationChecker.CanUserInvite(ctx, in.TeamId, in.InviterUserId)

	if err != nil {
		return nil, err
	}

	if !canInvite {
		return nil, fmt.Errorf("forbidden")
	}

	invited, err := e.memberInviter.InviteMember(ctx, in.UserId, in.TeamId, in.Role)

	if err != nil {
		return nil, err
	}

	if in.Email != "" {
		_ = e.emailSender.SendEmail(in.Email, simpleEmailText)
	}

	return &InviteResult{
		UserId: invited.UserId,
		TeamId: invited.TeamId,
		Role:   invited.Role,
	}, nil
}
