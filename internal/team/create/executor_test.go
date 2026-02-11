package create

import (
	"context"
	"errors"
	"testing"

	"mkk-luna-test-task/internal/team"

	"github.com/stretchr/testify/assert"
)

type stubTeamCreatorAndOwnerMakerSuccess struct{}

func (s *stubTeamCreatorAndOwnerMakerSuccess) CreateTeamAndMakeUserItsOwner(ctx context.Context, name string, userId int) (*team.Model, error) {
	return &team.Model{Id: 42, Name: name}, nil
}

type stubTeamCreatorAndOwnerMakerError struct{}

func (s *stubTeamCreatorAndOwnerMakerError) CreateTeamAndMakeUserItsOwner(ctx context.Context, name string, userId int) (*team.Model, error) {
	return nil, errors.New("db error")
}

func TestExecutor_Execute(t *testing.T) {
	type fields struct {
		teamCreatorAndOwnerMaker teamCreatorAndOwnerMaker
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
		expectedErr error
	}{
		{
			name: "success",
			fields: fields{
				teamCreatorAndOwnerMaker: &stubTeamCreatorAndOwnerMakerSuccess{},
			},
			args: args{
				ctx: context.Background(),
				in:  CreateInput{UserId: 101, Name: "New Team"},
			},
			want:    &CreateResult{Id: 42, Name: "New Team"},
			wantErr: false,
		},
		{
			name: "fail underlying",
			fields: fields{
				teamCreatorAndOwnerMaker: &stubTeamCreatorAndOwnerMakerError{},
			},
			args: args{
				ctx: context.Background(),
				in:  CreateInput{UserId: 505, Name: "Fail Team"},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				teamCreatorAndOwnerMaker: tt.fields.teamCreatorAndOwnerMaker,
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
