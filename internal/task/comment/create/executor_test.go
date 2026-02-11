package create

import (
	"context"
	"errors"
	"testing"
	"time"

	"mkk-luna-test-task/internal/task/comment"

	"github.com/stretchr/testify/assert"
)

type stubTaskCommentCreatorSuccess struct{}

func (s *stubTaskCommentCreatorSuccess) CreateTaskComment(ctx context.Context, commenterId int, taskId int, text string, createdAt time.Time) (*comment.Model, error) {
	return &comment.Model{
		Id:          42,
		TaskId:      taskId,
		CommenterId: commenterId,
		Text:        text,
		CreatedAt:   createdAt,
	}, nil
}

type stubTaskCommentCreatorError struct{}

func (s *stubTaskCommentCreatorError) CreateTaskComment(ctx context.Context, commenterId int, taskId int, text string, createdAt time.Time) (*comment.Model, error) {
	return nil, errors.New("db error")
}

type stubTeamMembershipCheckerMember struct{}

func (s *stubTeamMembershipCheckerMember) IsPartOfTeamByTaskId(userId, taskId int) (bool, error) {
	return true, nil
}

type stubTeamMembershipCheckerNotMember struct{}

func (s *stubTeamMembershipCheckerNotMember) IsPartOfTeamByTaskId(userId, taskId int) (bool, error) {
	return false, nil
}

type stubTeamMembershipCheckerError struct{}

func (s *stubTeamMembershipCheckerError) IsPartOfTeamByTaskId(userId, taskId int) (bool, error) {
	return false, errors.New("membership error")
}

func TestExecutor_Execute(t *testing.T) {
	type fields struct {
		taskCommentCreator    taskCommentCreator
		teamMembershipChecker teamMembershipChecker
	}
	type args struct {
		ctx context.Context
		in  CreateInput
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantId      int
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success - creates comment",
			fields: fields{
				taskCommentCreator:    &stubTaskCommentCreatorSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerMember{},
			},
			args: args{
				ctx: context.Background(),
				in: CreateInput{
					CommenterId: 1,
					TaskId:      2,
					Text:        "hi",
				},
			},
			wantId:  42,
			wantErr: false,
		},
		{
			name: "not a member - forbidden",
			fields: fields{
				taskCommentCreator:    &stubTaskCommentCreatorSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerNotMember{},
			},
			args: args{
				ctx: context.Background(),
				in: CreateInput{
					CommenterId: 1,
					TaskId:      2,
					Text:        "hi",
				},
			},
			wantId:      0,
			wantErr:     true,
			expectedErr: errors.New("forbidden"),
		},
		{
			name: "membership checker error",
			fields: fields{
				taskCommentCreator:    &stubTaskCommentCreatorSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerError{},
			},
			args: args{
				ctx: context.Background(),
				in: CreateInput{
					CommenterId: 1,
					TaskId:      2,
					Text:        "hi",
				},
			},
			wantId:      0,
			wantErr:     true,
			expectedErr: errors.New("membership error"),
		},
		{
			name: "comment creator error",
			fields: fields{
				taskCommentCreator:    &stubTaskCommentCreatorError{},
				teamMembershipChecker: &stubTeamMembershipCheckerMember{},
			},
			args: args{
				ctx: context.Background(),
				in: CreateInput{
					CommenterId: 6,
					TaskId:      7,
					Text:        "fail plz",
				},
			},
			wantId:      0,
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				taskCommentCreator:    tt.fields.taskCommentCreator,
				teamMembershipChecker: tt.fields.teamMembershipChecker,
			}

			gotId, err := e.Execute(tt.args.ctx, tt.args.in)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantId, gotId)

				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantId, gotId)
			}
		})
	}
}
