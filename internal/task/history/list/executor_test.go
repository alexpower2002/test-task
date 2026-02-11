package list

import (
	"context"
	"errors"
	"testing"
	"time"

	"mkk-luna-test-task/internal/task/history"

	"github.com/stretchr/testify/assert"
)

type stubHistoryListerSuccess struct {
	histories []*history.Model
}

func (s *stubHistoryListerSuccess) ListHistory(ctx context.Context, taskId int, startFromId int, limit int) ([]*history.Model, error) {
	return s.histories, nil
}

type stubHistoryListerEmpty struct{}

func (s *stubHistoryListerEmpty) ListHistory(ctx context.Context, taskId int, startFromId int, limit int) ([]*history.Model, error) {
	return []*history.Model{}, nil
}

type stubHistoryListerError struct{}

func (s *stubHistoryListerError) ListHistory(ctx context.Context, taskId int, startFromId int, limit int) ([]*history.Model, error) {
	return nil, errors.New("db error")
}

type stubTaskMembershipCheckerAllowed struct{}

func (s *stubTaskMembershipCheckerAllowed) IsPartOfTeamByTaskId(userId, taskId int) (bool, error) {
	return true, nil
}

type stubTaskMembershipCheckerDenied struct{}

func (s *stubTaskMembershipCheckerDenied) IsPartOfTeamByTaskId(userId, taskId int) (bool, error) {
	return false, nil
}

type stubTaskMembershipCheckerError struct{}

func (s *stubTaskMembershipCheckerError) IsPartOfTeamByTaskId(userId, taskId int) (bool, error) {
	return false, errors.New("membership check error")
}

func makeHistory(id int) *history.Model {
	return &history.Model{
		Id:          id,
		TaskId:      100,
		Status:      "done",
		Title:       "History Title",
		Description: "History Desc",
		CreatorId:   1,
		CreatedAt:   time.Unix(int64(id), 0),
		AssigneeId:  2,
		TeamId:      42,
		ChangedBy:   99,
		ChangedAt:   time.Unix(int64(id+1), 0),
	}
}

func TestExecutor_Execute(t *testing.T) {
	type fields struct {
		historyLister         historyLister
		taskMembershipChecker taskMembershipChecker
	}
	type args struct {
		ctx context.Context
		in  ListInput
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *ListResult
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success - single history",
			fields: fields{
				historyLister:         &stubHistoryListerSuccess{histories: []*history.Model{makeHistory(1)}},
				taskMembershipChecker: &stubTaskMembershipCheckerAllowed{},
			},
			args: args{
				ctx: context.Background(),
				in:  ListInput{UserId: 1, TaskId: 100, StartFromId: 0},
			},
			want: &ListResult{
				History: []HistoryItem{
					{
						Id:          1,
						TaskId:      100,
						Status:      "done",
						Title:       "History Title",
						Description: "History Desc",
						CreatorId:   1,
						CreatedAt:   time.Unix(1, 0),
						AssigneeId:  2,
						TeamId:      42,
						ChangedBy:   99,
						ChangedAt:   time.Unix(2, 0),
					},
				},
				NextStartFromId: 0,
				HasMore:         false,
			},
			wantErr: false,
		},
		{
			name: "success - no histories",
			fields: fields{
				historyLister:         &stubHistoryListerEmpty{},
				taskMembershipChecker: &stubTaskMembershipCheckerAllowed{},
			},
			args: args{
				ctx: context.Background(),
				in:  ListInput{UserId: 1, TaskId: 100, StartFromId: 0},
			},
			want: &ListResult{
				History:         []HistoryItem{},
				NextStartFromId: 0,
				HasMore:         false,
			},
			wantErr: false,
		},
		{
			name: "success - hasMore true",
			fields: fields{
				historyLister: func() historyLister {
					histories := make([]*history.Model, 0, defaultLimit+1)
					for i := 1; i <= defaultLimit+1; i++ {
						histories = append(histories, makeHistory(i))
					}
					return &stubHistoryListerSuccess{histories: histories}
				}(),
				taskMembershipChecker: &stubTaskMembershipCheckerAllowed{},
			},
			args: args{
				ctx: context.Background(),
				in:  ListInput{UserId: 1, TaskId: 100, StartFromId: 0},
			},
			want: func() *ListResult {
				historyItems := make([]HistoryItem, 0, defaultLimit)
				for i := 1; i <= defaultLimit; i++ {
					model := makeHistory(i)
					historyItems = append(historyItems, HistoryItem{
						Id:          model.Id,
						TaskId:      model.TaskId,
						Status:      model.Status,
						Title:       model.Title,
						Description: model.Description,
						CreatorId:   model.CreatorId,
						CreatedAt:   model.CreatedAt,
						AssigneeId:  model.AssigneeId,
						TeamId:      model.TeamId,
						ChangedBy:   model.ChangedBy,
						ChangedAt:   model.ChangedAt,
					})
				}
				return &ListResult{
					History:         historyItems,
					NextStartFromId: defaultLimit + 1,
					HasMore:         true,
				}
			}(),
			wantErr: false,
		},
		{
			name: "fail - forbidden by team membership",
			fields: fields{
				historyLister:         &stubHistoryListerSuccess{histories: []*history.Model{makeHistory(1)}},
				taskMembershipChecker: &stubTaskMembershipCheckerDenied{},
			},
			args: args{
				ctx: context.Background(),
				in:  ListInput{UserId: 999, TaskId: 100, StartFromId: 0},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("forbidden"),
		},
		{
			name: "fail - error checking team membership",
			fields: fields{
				historyLister:         &stubHistoryListerSuccess{histories: []*history.Model{makeHistory(1)}},
				taskMembershipChecker: &stubTaskMembershipCheckerError{},
			},
			args: args{
				ctx: context.Background(),
				in:  ListInput{UserId: 888, TaskId: 777, StartFromId: 0},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("failed to check team membership: membership check error"),
		},
		{
			name: "fail - lister returns error",
			fields: fields{
				historyLister:         &stubHistoryListerError{},
				taskMembershipChecker: &stubTaskMembershipCheckerAllowed{},
			},
			args: args{
				ctx: context.Background(),
				in:  ListInput{UserId: 123, TaskId: 98, StartFromId: 0},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				historyLister:         tt.fields.historyLister,
				taskMembershipChecker: tt.fields.taskMembershipChecker,
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
