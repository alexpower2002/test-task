package invite

import (
	"context"
	"errors"
	"testing"

	"mkk-luna-test-task/internal/team/member"

	"github.com/stretchr/testify/assert"
)

type stubMemberInviterSuccess struct{}

func (s *stubMemberInviterSuccess) InviteMember(ctx context.Context, userId, teamId int, role member.Role) (*member.Model, error) {
	return &member.Model{
		UserId: userId,
		TeamId: teamId,
		Role:   role,
	}, nil
}

type stubMemberInviterError struct{}

func (s *stubMemberInviterError) InviteMember(ctx context.Context, userId, teamId int, role member.Role) (*member.Model, error) {
	return nil, errors.New("db error")
}

type stubInvitationCheckerAllow struct{}

func (s *stubInvitationCheckerAllow) CanUserInvite(ctx context.Context, teamId int, userId int) (bool, error) {
	return true, nil
}

type stubInvitationCheckerForbid struct{}

func (s *stubInvitationCheckerForbid) CanUserInvite(ctx context.Context, teamId int, userId int) (bool, error) {
	return false, nil
}

type stubInvitationCheckerError struct{}

func (s *stubInvitationCheckerError) CanUserInvite(ctx context.Context, teamId int, userId int) (bool, error) {
	return false, errors.New("perms check error")
}

type stubEmailSender struct{}

func (s *stubEmailSender) SendEmail(address string, text string) error {
	return nil
}

func TestExecutor_Execute(t *testing.T) {
	type fields struct {
		memberInviter           memberInviter
		memberInvitationChecker memberInvitationChecker
	}
	type args struct {
		ctx context.Context
		in  InviteInput
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *InviteResult
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success",
			fields: fields{
				memberInviter:           &stubMemberInviterSuccess{},
				memberInvitationChecker: &stubInvitationCheckerAllow{},
			},
			args: args{
				ctx: context.Background(),
				in: InviteInput{
					InviterUserId: 1,
					TeamId:        10,
					UserId:        20,
					Role:          member.NormalRole,
				},
			},
			want: &InviteResult{
				UserId: 20,
				TeamId: 10,
				Role:   member.NormalRole,
			},
			wantErr: false,
		},
		{
			name: "invitation forbidden",
			fields: fields{
				memberInviter:           &stubMemberInviterSuccess{},
				memberInvitationChecker: &stubInvitationCheckerForbid{},
			},
			args: args{
				ctx: context.Background(),
				in: InviteInput{
					InviterUserId: 2,
					TeamId:        20,
					UserId:        30,
					Role:          member.NormalRole,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("forbidden"),
		},
		{
			name: "invitation checker error",
			fields: fields{
				memberInviter:           &stubMemberInviterSuccess{},
				memberInvitationChecker: &stubInvitationCheckerError{},
			},
			args: args{
				ctx: context.Background(),
				in: InviteInput{
					InviterUserId: 3,
					TeamId:        30,
					UserId:        40,
					Role:          member.NormalRole,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("perms check error"),
		},
		{
			name: "member inviter error",
			fields: fields{
				memberInviter:           &stubMemberInviterError{},
				memberInvitationChecker: &stubInvitationCheckerAllow{},
			},
			args: args{
				ctx: context.Background(),
				in: InviteInput{
					InviterUserId: 4,
					TeamId:        40,
					UserId:        50,
					Role:          member.NormalRole,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				memberInviter:           tt.fields.memberInviter,
				memberInvitationChecker: tt.fields.memberInvitationChecker,
				emailSender:             &stubEmailSender{},
			}

			got, err := e.Execute(tt.args.ctx, tt.args.in)

			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)

				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
