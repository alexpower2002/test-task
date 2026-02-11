package list

import (
	"context"
	"errors"
	"testing"

	"mkk-luna-test-task/internal/task"

	"github.com/stretchr/testify/assert"
)

type stubTaskListerSuccess struct{}

func (s *stubTaskListerSuccess) ListTasks(ctx context.Context, teamId int, status string, assigneeId int, startFromId int, limit int) ([]*task.Model, error) {
	return []*task.Model{
		{
			Id:          1,
			Status:      "open",
			Title:       "Test Task 1",
			Description: "Description 1",
			CreatorId:   10,
			AssigneeId:  20,
			TeamId:      teamId,
		},
		{
			Id:          2,
			Status:      "in_progress",
			Title:       "Test Task 2",
			Description: "Description 2",
			CreatorId:   11,
			AssigneeId:  21,
			TeamId:      teamId,
		},
	}, nil
}

type stubTaskListerEmpty struct{}

func (s *stubTaskListerEmpty) ListTasks(ctx context.Context, teamId int, status string, assigneeId int, startFromId int, limit int) ([]*task.Model, error) {
	return []*task.Model{}, nil
}

type stubTaskListerError struct{}

func (s *stubTaskListerError) ListTasks(ctx context.Context, teamId int, status string, assigneeId int, startFromId int, limit int) ([]*task.Model, error) {
	return nil, errors.New("db error")
}

type stubTeamMembershipCheckerAllow struct{}

func (s *stubTeamMembershipCheckerAllow) CheckUserIsInTeam(userId, teamId int) (bool, error) {
	return true, nil
}

type stubTeamMembershipCheckerDeny struct{}

func (s *stubTeamMembershipCheckerDeny) CheckUserIsInTeam(userId, teamId int) (bool, error) {
	return false, nil
}

type stubTeamMembershipCheckerError struct{}

func (s *stubTeamMembershipCheckerError) CheckUserIsInTeam(userId, teamId int) (bool, error) {
	return false, errors.New("membership check error")
}

type mockTaskCacheLister struct {
	tasks    []task.Model
	hit      bool
	lastArgs struct {
		ctx         context.Context
		teamId      int
		limit       int
		startFromID *int
	}
}

func (s *mockTaskCacheLister) ListTasksFromCache(ctx context.Context, teamId int, limit int, startFromID *int) ([]task.Model, bool, error) {
	s.lastArgs.ctx = ctx
	s.lastArgs.teamId = teamId
	s.lastArgs.limit = limit
	s.lastArgs.startFromID = startFromID

	return s.tasks, s.hit, nil
}

type mockTaskCacheWriter struct {
	tasksWritten []task.Model
	teamId       int
	called       bool
}

func (c *mockTaskCacheWriter) CacheTasks(ctx context.Context, teamId int, tasks []task.Model) error {
	c.called = true
	c.tasksWritten = tasks
	c.teamId = teamId

	return nil
}

func TestExecutor_Execute(t *testing.T) {
	type fields struct {
		taskLister            taskLister
		teamMembershipChecker teamMembershipChecker
		cacheLister           taskCacheLister
		cacheWriter           taskCacheWriter
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
		expectCache bool
	}{
		{
			name: "success - multiple tasks",
			fields: fields{
				taskLister:            &stubTaskListerSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerAllow{},
				cacheLister:           &mockTaskCacheLister{},
				cacheWriter:           &mockTaskCacheWriter{},
			},
			args: args{
				ctx: context.Background(),
				in: ListInput{
					UserId:      100,
					TeamId:      1,
					Status:      "",
					AssigneeId:  0,
					StartFromId: 0,
				},
			},
			want: &ListResult{
				Tasks: []TaskItem{
					{
						Id:          1,
						Status:      "open",
						Title:       "Test Task 1",
						Description: "Description 1",
						CreatorId:   10,
						AssigneeId:  20,
						TeamId:      1,
					},
					{
						Id:          2,
						Status:      "in_progress",
						Title:       "Test Task 2",
						Description: "Description 2",
						CreatorId:   11,
						AssigneeId:  21,
						TeamId:      1,
					},
				},
				NextStartFromId: 0,
				HasMore:         false,
			},
			wantErr: false,
		},
		{
			name: "success - no tasks",
			fields: fields{
				taskLister:            &stubTaskListerEmpty{},
				teamMembershipChecker: &stubTeamMembershipCheckerAllow{},
				cacheLister:           &mockTaskCacheLister{},
				cacheWriter:           &mockTaskCacheWriter{},
			},
			args: args{
				ctx: context.Background(),
				in: ListInput{
					UserId:      101,
					TeamId:      1,
					Status:      "",
					AssigneeId:  0,
					StartFromId: 0,
				},
			},
			want: &ListResult{
				Tasks:           []TaskItem{},
				NextStartFromId: 0,
				HasMore:         false,
			},
			wantErr: false,
		},
		{
			name: "fail - forbidden by team membership",
			fields: fields{
				taskLister:            &stubTaskListerSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerDeny{},
				cacheLister:           &mockTaskCacheLister{},
				cacheWriter:           &mockTaskCacheWriter{},
			},
			args: args{
				ctx: context.Background(),
				in: ListInput{
					UserId:      999,
					TeamId:      8,
					Status:      "",
					AssigneeId:  0,
					StartFromId: 0,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("forbidden"),
		},
		{
			name: "fail - error checking team membership",
			fields: fields{
				taskLister:            &stubTaskListerSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerError{},
				cacheLister:           &mockTaskCacheLister{},
				cacheWriter:           &mockTaskCacheWriter{},
			},
			args: args{
				ctx: context.Background(),
				in: ListInput{
					UserId:      888,
					TeamId:      7,
					Status:      "",
					AssigneeId:  0,
					StartFromId: 0,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("membership check error"),
		},
		{
			name: "fail - lister returns error",
			fields: fields{
				taskLister:            &stubTaskListerError{},
				teamMembershipChecker: &stubTeamMembershipCheckerAllow{},
				cacheLister:           &mockTaskCacheLister{},
				cacheWriter:           &mockTaskCacheWriter{},
			},
			args: args{
				ctx: context.Background(),
				in: ListInput{
					UserId:      123,
					TeamId:      3,
					Status:      "",
					AssigneeId:  0,
					StartFromId: 0,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
		{
			name: "success - cache hit returns tasks",
			fields: fields{
				taskLister:            &stubTaskListerSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerAllow{},
				cacheLister: &mockTaskCacheLister{
					tasks: []task.Model{
						{
							Id: 101, Status: "open", Title: "Cached Task 1", Description: "DescA", CreatorId: 1, AssigneeId: 2, TeamId: 5,
						},
						{
							Id: 102, Status: "closed", Title: "Cached Task 2", Description: "DescB", CreatorId: 3, AssigneeId: 4, TeamId: 5,
						},
					},
					hit: true,
				},
				cacheWriter: &mockTaskCacheWriter{},
			},
			args: args{
				ctx: context.Background(),
				in: ListInput{
					UserId:      1,
					TeamId:      5,
					Status:      "",
					AssigneeId:  0,
					StartFromId: 0,
				},
			},
			want: &ListResult{
				Tasks: []TaskItem{
					{
						Id: 101, Status: "open", Title: "Cached Task 1", Description: "DescA", CreatorId: 1, AssigneeId: 2, TeamId: 5,
					},
					{
						Id: 102, Status: "closed", Title: "Cached Task 2", Description: "DescB", CreatorId: 3, AssigneeId: 4, TeamId: 5,
					},
				},
				NextStartFromId: 0,
				HasMore:         false,
			},
			wantErr:     false,
			expectCache: true,
		},

		{
			name: "success - cache miss falls back to DB and writes to cache",
			fields: func() fields {
				writer := &mockTaskCacheWriter{}
				return fields{
					taskLister:            &stubTaskListerSuccess{},
					teamMembershipChecker: &stubTeamMembershipCheckerAllow{},
					cacheLister: &mockTaskCacheLister{
						tasks: nil, hit: false,
					},
					cacheWriter: writer,
				}
			}(),
			args: args{
				ctx: context.Background(),
				in: ListInput{
					UserId:      222,
					TeamId:      77,
					Status:      "",
					AssigneeId:  0,
					StartFromId: 0,
				},
			},
			want: &ListResult{
				Tasks: []TaskItem{
					{
						Id:          1,
						Status:      "open",
						Title:       "Test Task 1",
						Description: "Description 1",
						CreatorId:   10,
						AssigneeId:  20,
						TeamId:      77,
					},
					{
						Id:          2,
						Status:      "in_progress",
						Title:       "Test Task 2",
						Description: "Description 2",
						CreatorId:   11,
						AssigneeId:  21,
						TeamId:      77,
					},
				},
				NextStartFromId: 0,
				HasMore:         false,
			},
			wantErr:     false,
			expectCache: false,
		},
		{
			name: "success - does not use cache if Status specified (search)",
			fields: fields{
				taskLister:            &stubTaskListerSuccess{},
				teamMembershipChecker: &stubTeamMembershipCheckerAllow{},
				cacheLister: &mockTaskCacheLister{
					tasks: nil, hit: false,
				},
				cacheWriter: &mockTaskCacheWriter{},
			},
			args: args{
				ctx: context.Background(),
				in: ListInput{
					UserId:      333,
					TeamId:      42,
					Status:      "open",
					AssigneeId:  0,
					StartFromId: 0,
				},
			},
			want: &ListResult{
				Tasks: []TaskItem{
					{
						Id:          1,
						Status:      "open",
						Title:       "Test Task 1",
						Description: "Description 1",
						CreatorId:   10,
						AssigneeId:  20,
						TeamId:      42,
					},
					{
						Id:          2,
						Status:      "in_progress",
						Title:       "Test Task 2",
						Description: "Description 2",
						CreatorId:   11,
						AssigneeId:  21,
						TeamId:      42,
					},
				},
				NextStartFromId: 0,
				HasMore:         false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				taskLister:            tt.fields.taskLister,
				teamMembershipChecker: tt.fields.teamMembershipChecker,
				cacheLister:           tt.fields.cacheLister,
				cacheWriter:           tt.fields.cacheWriter,
			}

			got, err := e.Execute(tt.args.ctx, tt.args.in)

			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)

				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}

				return
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			writer, _ := tt.fields.cacheWriter.(*mockTaskCacheWriter)

			if tt.expectCache {
				assert.False(t, writer.called)
			} else {
				cacheLister, _ := tt.fields.cacheLister.(*mockTaskCacheLister)

				if !cacheLister.hit && tt.args.in.Status == "" {
					assert.True(t, writer.called)
					assert.Equal(t, tt.args.in.TeamId, writer.teamId)

					if len(tt.want.Tasks) > 0 {
						assert.NotEmpty(t, writer.tasksWritten)
					}
				} else {
					assert.False(t, writer.called)
				}
			}
		})
	}
}
