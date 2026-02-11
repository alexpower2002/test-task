package create

import (
	"context"
	"errors"
	"testing"
	"time"

	"mkk-luna-test-task/internal/task"

	"github.com/stretchr/testify/assert"
)

type stubTaskCreatorSuccess struct{}

func (s *stubTaskCreatorSuccess) CreateTask(ctx context.Context, status, title, description string, creatorId, assigneeId, teamId int, createdAt time.Time) (*task.Model, error) {
	return &task.Model{
		Id:          11,
		Status:      status,
		Title:       title,
		Description: description,
		CreatorId:   creatorId,
		CreatedAt:   createdAt,
		AssigneeId:  assigneeId,
		TeamId:      teamId,
	}, nil
}

type stubTaskCreatorError struct{}

func (s *stubTaskCreatorError) CreateTask(ctx context.Context, status, title, description string, creatorId, assigneeId, teamId int, createdAt time.Time) (*task.Model, error) {
	return nil, errors.New("some creation error")
}

type stubTeamMembershipCheckerAllowed struct{}

func (s *stubTeamMembershipCheckerAllowed) CheckUserIsInTeam(userId, teamId int) (bool, error) {
	return true, nil
}

type stubTeamMembershipCheckerForbidden struct{}

func (s *stubTeamMembershipCheckerForbidden) CheckUserIsInTeam(userId, teamId int) (bool, error) {
	return false, nil
}

type stubTeamMembershipCheckerError struct{}

func (s *stubTeamMembershipCheckerError) CheckUserIsInTeam(userId, teamId int) (bool, error) {
	return false, errors.New("membership check failed")
}

func TestExecutor_Execute(t *testing.T) {
	type fields struct {
		taskCreator           taskCreator
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
		want        *CreateResult
		wantErr     bool
		expectedErr string
	}{
		{
			name: "success",
			fields: fields{
				taskCreator:           &stubTaskCreatorSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerAllowed{},
			},
			args: args{
				ctx: context.Background(),
				in: CreateInput{
					CreatorId:   1001,
					Status:      "open",
					Title:       "Test Title",
					Description: "Test Description",
					AssigneeId:  2002,
					TeamId:      3003,
				},
			},
			want: &CreateResult{
				Id:          11,
				Status:      "open",
				Title:       "Test Title",
				Description: "Test Description",
				CreatorId:   1001,
				AssigneeId:  2002,
				TeamId:      3003,
			},
			wantErr: false,
		},
		{
			name: "forbidden membership",
			fields: fields{
				taskCreator:           &stubTaskCreatorSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerForbidden{},
			},
			args: args{
				ctx: context.Background(),
				in: CreateInput{
					CreatorId:   1,
					Status:      "open",
					Title:       "Try",
					Description: "Try Desc",
					AssigneeId:  2,
					TeamId:      3,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: "forbidden",
		},
		{
			name: "error in membership checker",
			fields: fields{
				taskCreator:           &stubTaskCreatorSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerError{},
			},
			args: args{
				ctx: context.Background(),
				in: CreateInput{
					CreatorId:   8,
					Status:      "X",
					Title:       "Y",
					Description: "Z",
					AssigneeId:  13,
					TeamId:      21,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: "membership check failed",
		},
		{
			name: "creation fails",
			fields: fields{
				taskCreator:           &stubTaskCreatorError{},
				teamMembershipChecker: &stubTeamMembershipCheckerAllowed{},
			},
			args: args{
				ctx: context.Background(),
				in: CreateInput{
					CreatorId:   404,
					Status:      "fail",
					Title:       "nope",
					Description: "should fail",
					AssigneeId:  505,
					TeamId:      606,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: "some creation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				taskCreator:           tt.fields.taskCreator,
				teamMembershipChecker: tt.fields.teamMembershipChecker,
			}

			got, err := e.Execute(tt.args.ctx, tt.args.in)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)

				if tt.expectedErr != "" {
					assert.EqualError(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.Id, got.Id)
				assert.Equal(t, tt.want.Status, got.Status)
				assert.Equal(t, tt.want.Title, got.Title)
				assert.Equal(t, tt.want.Description, got.Description)
				assert.Equal(t, tt.want.CreatorId, got.CreatorId)
				assert.Equal(t, tt.want.AssigneeId, got.AssigneeId)
				assert.Equal(t, tt.want.TeamId, got.TeamId)
				assert.False(t, got.CreatedAt.IsZero())
			}
		})
	}
}
