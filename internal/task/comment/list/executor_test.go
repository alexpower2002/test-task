package list

import (
	"context"
	"errors"
	"testing"
	"time"

	"mkk-luna-test-task/internal/task/comment"

	"github.com/stretchr/testify/assert"
)

type stubTaskCommentListerSuccess struct{}

func (s *stubTaskCommentListerSuccess) ListTaskComments(ctx context.Context, taskId int, startFromId int, limit int) ([]comment.Model, error) {
	return []comment.Model{
		{
			Id:          1,
			TaskId:      taskId,
			CommenterId: 5,
			Text:        "test comment 1",
			CreatedAt:   time.Date(2023, 4, 1, 13, 0, 0, 0, time.UTC),
		},
		{
			Id:          2,
			TaskId:      taskId,
			CommenterId: 6,
			Text:        "test comment 2",
			CreatedAt:   time.Date(2023, 4, 2, 14, 1, 0, 0, time.UTC),
		},
	}, nil
}

type stubTaskCommentListerPaginate struct{}

func (s *stubTaskCommentListerPaginate) ListTaskComments(ctx context.Context, taskId int, startFromId int, limit int) ([]comment.Model, error) {
	models := []comment.Model{}
	for i := 0; i < limit+1; i++ {
		models = append(models, comment.Model{
			Id:          i + 100,
			TaskId:      taskId,
			CommenterId: 10 + i,
			Text:        "msg",
			CreatedAt:   time.Now(),
		})
	}
	return models, nil
}

type stubTaskCommentListerEmpty struct{}

func (s *stubTaskCommentListerEmpty) ListTaskComments(ctx context.Context, taskId int, startFromId int, limit int) ([]comment.Model, error) {
	return []comment.Model{}, nil
}

type stubTaskCommentListerError struct{}

func (s *stubTaskCommentListerError) ListTaskComments(ctx context.Context, taskId int, startFromId int, limit int) ([]comment.Model, error) {
	return nil, errors.New("db error")
}

func TestExecutor_Execute(t *testing.T) {
	type fields struct {
		taskCommentLister taskCommentLister
	}
	type args struct {
		ctx         context.Context
		taskId      int
		startFromId int
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantResult      *ListResult
		wantHasMore     bool
		wantNextFromId  int
		wantErr         bool
		expectedErr     error
		wantCommentsLen int
	}{
		{
			name: "success - returns comments",
			fields: fields{
				taskCommentLister: &stubTaskCommentListerSuccess{},
			},
			args: args{
				ctx:         context.Background(),
				taskId:      42,
				startFromId: 0,
			},
			wantHasMore:     false,
			wantNextFromId:  0,
			wantErr:         false,
			wantCommentsLen: 2,
		},
		{
			name: "success - paginated with hasMore",
			fields: fields{
				taskCommentLister: &stubTaskCommentListerPaginate{},
			},
			args: args{
				ctx:         context.Background(),
				taskId:      1,
				startFromId: 0,
			},
			wantHasMore:     true,
			wantNextFromId:  200,
			wantErr:         false,
			wantCommentsLen: 100,
		},
		{
			name: "success - empty result",
			fields: fields{
				taskCommentLister: &stubTaskCommentListerEmpty{},
			},
			args: args{
				ctx:         context.Background(),
				taskId:      100,
				startFromId: 0,
			},
			wantHasMore:     false,
			wantNextFromId:  0,
			wantErr:         false,
			wantCommentsLen: 0,
		},
		{
			name: "error - lister returns error",
			fields: fields{
				taskCommentLister: &stubTaskCommentListerError{},
			},
			args: args{
				ctx:         context.Background(),
				taskId:      1,
				startFromId: 0,
			},
			wantHasMore:     false,
			wantNextFromId:  0,
			wantErr:         true,
			expectedErr:     errors.New("db error"),
			wantCommentsLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				taskCommentLister: tt.fields.taskCommentLister,
			}

			got, err := e.Execute(tt.args.ctx, tt.args.taskId, tt.args.startFromId)

			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)

				if tt.expectedErr != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantHasMore, got.HasMore)
				assert.Equal(t, tt.wantNextFromId, got.NextStartFromId)
				assert.Len(t, got.Comments, tt.wantCommentsLen)
			}
		})
	}
}
