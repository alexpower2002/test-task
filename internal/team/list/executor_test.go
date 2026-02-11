package list

import (
	"context"
	"errors"
	"testing"

	"mkk-luna-test-task/internal/team"

	"github.com/stretchr/testify/assert"
)

type stubTeamListerSuccess struct{}

func (s *stubTeamListerSuccess) ListTeams(ctx context.Context, userId int) ([]*team.Model, error) {
	return []*team.Model{
		{Id: 1, Name: "Alpha"},
		{Id: 2, Name: "Beta"},
	}, nil
}

type stubTeamListerEmpty struct{}

func (s *stubTeamListerEmpty) ListTeams(ctx context.Context, userId int) ([]*team.Model, error) {
	return []*team.Model{}, nil
}

type stubTeamListerError struct{}

func (s *stubTeamListerError) ListTeams(ctx context.Context, userId int) ([]*team.Model, error) {
	return nil, errors.New("db error")
}

func TestExecutor_Execute(t *testing.T) {
	type fields struct {
		teamLister teamLister
	}
	type args struct {
		ctx    context.Context
		userId int
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
			name: "success - multiple teams",
			fields: fields{
				teamLister: &stubTeamListerSuccess{},
			},
			args: args{
				ctx:    context.Background(),
				userId: 123,
			},
			want: &ListResult{
				Teams: []TeamItem{
					{Id: 1, Name: "Alpha"},
					{Id: 2, Name: "Beta"},
				},
			},
			wantErr: false,
		},
		{
			name: "success - no teams",
			fields: fields{
				teamLister: &stubTeamListerEmpty{},
			},
			args: args{
				ctx:    context.Background(),
				userId: 999,
			},
			want: &ListResult{
				Teams: []TeamItem{},
			},
			wantErr: false,
		},
		{
			name: "error from lister",
			fields: fields{
				teamLister: &stubTeamListerError{},
			},
			args: args{
				ctx:    context.Background(),
				userId: 333,
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				teamLister: tt.fields.teamLister,
			}

			got, err := e.Execute(tt.args.ctx, tt.args.userId)

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
