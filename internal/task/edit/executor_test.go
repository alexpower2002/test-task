package edit

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"mkk-luna-test-task/internal/task"
	"mkk-luna-test-task/internal/task/history"

	"github.com/stretchr/testify/assert"
)

type stubTaskEditorSuccess struct{}

func (s *stubTaskEditorSuccess) EditTask(ctx context.Context, id int, status, title, description string, assigneeId int) error {
	return nil
}

type stubTaskEditorError struct{}

func (s *stubTaskEditorError) EditTask(ctx context.Context, id int, status, title, description string, assigneeId int) error {
	return errors.New("edit task error")
}

type stubTaskGetterSuccess struct {
	task *task.Model
}

func (s *stubTaskGetterSuccess) GetTaskById(ctx context.Context, id int) (*task.Model, error) {
	return s.task, nil
}

type stubTaskGetterNotFound struct{}

func (s *stubTaskGetterNotFound) GetTaskById(ctx context.Context, id int) (*task.Model, error) {
	return nil, sql.ErrNoRows
}

type stubTaskGetterError struct{}

func (s *stubTaskGetterError) GetTaskById(ctx context.Context, id int) (*task.Model, error) {
	return nil, errors.New("db error")
}

type stubHistorySaverSuccess struct{}

func (s *stubHistorySaverSuccess) CreateHistory(ctx context.Context, t *task.Model, changedBy int, changedAt time.Time) (*history.Model, error) {
	return &history.Model{}, nil
}

type stubHistorySaverError struct{}

func (s *stubHistorySaverError) CreateHistory(ctx context.Context, t *task.Model, changedBy int, changedAt time.Time) (*history.Model, error) {
	return nil, errors.New("history error")
}

type stubTaskMembershipCheckerAllowed struct{}

func (s *stubTaskMembershipCheckerAllowed) IsPartOfTeamByTaskId(userId, taskId int) (bool, error) {
	return true, nil
}

type stubTaskMembershipCheckerForbidden struct{}

func (s *stubTaskMembershipCheckerForbidden) IsPartOfTeamByTaskId(userId, taskId int) (bool, error) {
	return false, nil
}

type stubTaskMembershipCheckerError struct{}

func (s *stubTaskMembershipCheckerError) IsPartOfTeamByTaskId(userId, taskId int) (bool, error) {
	return false, errors.New("membership error")
}

type mockCacheUpdater struct {
	called bool
}

func (m *mockCacheUpdater) UpdateTaskInCache(ctx context.Context, t task.Model) error {
	m.called = true
	return nil
}

func TestExecutor_Execute(t *testing.T) {
	type fields struct {
		taskEditor            taskEditor
		taskGetter            taskGetter
		historySaver          historySaver
		taskMembershipChecker taskMembershipChecker
		cacheUpdater          *mockCacheUpdater
	}
	type args struct {
		ctx context.Context
		in  EditInput
	}
	baseTask := &task.Model{
		Id:          1,
		Status:      "old",
		Title:       "old title",
		Description: "old desc",
		AssigneeId:  7,
		TeamId:      10,
		CreatorId:   100,
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		want        bool
		wantErr     bool
		expectedErr string
	}{
		{
			name: "success",
			fields: fields{
				taskEditor:            &stubTaskEditorSuccess{},
				taskGetter:            &stubTaskGetterSuccess{task: baseTask},
				historySaver:          &stubHistorySaverSuccess{},
				taskMembershipChecker: &stubTaskMembershipCheckerAllowed{},
				cacheUpdater:          &mockCacheUpdater{},
			},
			args: args{
				ctx: context.Background(),
				in: EditInput{
					UserId:      5,
					TaskId:      1,
					Status:      "inprogress",
					Title:       "new",
					Description: "desc",
					AssigneeId:  8,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "task not found",
			fields: fields{
				taskEditor:            &stubTaskEditorSuccess{},
				taskGetter:            &stubTaskGetterNotFound{},
				historySaver:          &stubHistorySaverSuccess{},
				taskMembershipChecker: &stubTaskMembershipCheckerAllowed{},
				cacheUpdater:          &mockCacheUpdater{},
			},
			args: args{
				ctx: context.Background(),
				in:  EditInput{TaskId: 42, UserId: 1},
			},
			want:        false,
			wantErr:     true,
			expectedErr: "task not found",
		},
		{
			name: "get task db error",
			fields: fields{
				taskEditor:            &stubTaskEditorSuccess{},
				taskGetter:            &stubTaskGetterError{},
				historySaver:          &stubHistorySaverSuccess{},
				taskMembershipChecker: &stubTaskMembershipCheckerAllowed{},
				cacheUpdater:          &mockCacheUpdater{},
			},
			args: args{
				ctx: context.Background(),
				in:  EditInput{TaskId: 13, UserId: 1},
			},
			want:        false,
			wantErr:     true,
			expectedErr: "db error",
		},
		{
			name: "membership forbidden",
			fields: fields{
				taskEditor:            &stubTaskEditorSuccess{},
				taskGetter:            &stubTaskGetterSuccess{task: baseTask},
				historySaver:          &stubHistorySaverSuccess{},
				taskMembershipChecker: &stubTaskMembershipCheckerForbidden{},
				cacheUpdater:          &mockCacheUpdater{},
			},
			args: args{
				ctx: context.Background(),
				in:  EditInput{TaskId: 1, UserId: 11},
			},
			want:        false,
			wantErr:     true,
			expectedErr: "forbidden",
		},
		{
			name: "membership checker error",
			fields: fields{
				taskEditor:            &stubTaskEditorSuccess{},
				taskGetter:            &stubTaskGetterSuccess{task: baseTask},
				historySaver:          &stubHistorySaverSuccess{},
				taskMembershipChecker: &stubTaskMembershipCheckerError{},
				cacheUpdater:          &mockCacheUpdater{},
			},
			args: args{
				ctx: context.Background(),
				in:  EditInput{TaskId: 1, UserId: 99},
			},
			want:        false,
			wantErr:     true,
			expectedErr: "failed to check team membership: membership error",
		},
		{
			name: "history saver error",
			fields: fields{
				taskEditor:            &stubTaskEditorSuccess{},
				taskGetter:            &stubTaskGetterSuccess{task: baseTask},
				historySaver:          &stubHistorySaverError{},
				taskMembershipChecker: &stubTaskMembershipCheckerAllowed{},
				cacheUpdater:          &mockCacheUpdater{},
			},
			args: args{
				ctx: context.Background(),
				in:  EditInput{TaskId: 1, UserId: 5},
			},
			want:        false,
			wantErr:     true,
			expectedErr: "failed to save task history: history error",
		},
		{
			name: "edit task error",
			fields: fields{
				taskEditor:            &stubTaskEditorError{},
				taskGetter:            &stubTaskGetterSuccess{task: baseTask},
				historySaver:          &stubHistorySaverSuccess{},
				taskMembershipChecker: &stubTaskMembershipCheckerAllowed{},
				cacheUpdater:          &mockCacheUpdater{},
			},
			args: args{
				ctx: context.Background(),
				in:  EditInput{TaskId: 1, UserId: 5},
			},
			want:        false,
			wantErr:     true,
			expectedErr: "edit task error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				taskEditor:            tt.fields.taskEditor,
				taskGetter:            tt.fields.taskGetter,
				historySaver:          tt.fields.historySaver,
				taskMembershipChecker: tt.fields.taskMembershipChecker,
				cacheUpdater:          tt.fields.cacheUpdater,
			}

			got, err := e.Execute(tt.args.ctx, tt.args.in)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, got)

				if tt.expectedErr != "" {
					assert.EqualError(t, err, tt.expectedErr)
				}

				assert.False(t, tt.fields.cacheUpdater.called, "expected cacheUpdater not to be called")
			} else {
				assert.NoError(t, err)
				assert.True(t, got)
				assert.True(t, tt.fields.cacheUpdater.called, "expected cacheUpdater to be called")
			}
		})
	}
}
