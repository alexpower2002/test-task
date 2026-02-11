package register

import (
	"context"
	"errors"
	"testing"

	"mkk-luna-test-task/internal/user"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type stubUserRegistererSuccess struct {
	returnUser *user.Model
}

func (s *stubUserRegistererSuccess) RegisterUser(ctx context.Context, username, passwordHashed string) (*user.Model, error) {
	return s.returnUser, nil
}

type stubUserRegistererFail struct {
	err error
}

func (s *stubUserRegistererFail) RegisterUser(ctx context.Context, username, passwordHashed string) (*user.Model, error) {
	return nil, s.err
}

func hashedPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func TestExecutor_Execute(t *testing.T) {
	validUsername := "newuser"
	validPassword := "securepass"
	createdUser := &user.Model{Id: 99, Username: validUsername, PasswordHashed: hashedPassword(validPassword)}

	type fields struct {
		userRegisterer userRegisterer
	}
	type args struct {
		ctx context.Context
		in  RegisterInput
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *RegisterResult
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success registration",
			fields: fields{
				userRegisterer: &stubUserRegistererSuccess{returnUser: createdUser},
			},
			args: args{
				ctx: context.Background(),
				in: RegisterInput{
					Username: validUsername,
					Password: validPassword,
				},
			},
			want:    &RegisterResult{UserId: createdUser.Id},
			wantErr: false,
		},
		{
			name: "missing username",
			fields: fields{
				userRegisterer: &stubUserRegistererSuccess{returnUser: createdUser},
			},
			args: args{
				ctx: context.Background(),
				in: RegisterInput{
					Username: "",
					Password: validPassword,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("username and password required"),
		},
		{
			name: "missing password",
			fields: fields{
				userRegisterer: &stubUserRegistererSuccess{returnUser: createdUser},
			},
			args: args{
				ctx: context.Background(),
				in: RegisterInput{
					Username: validUsername,
					Password: "",
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("username and password required"),
		},
		{
			name: "register user fails (database error)",
			fields: fields{
				userRegisterer: &stubUserRegistererFail{err: errors.New("db error")},
			},
			args: args{
				ctx: context.Background(),
				in: RegisterInput{
					Username: validUsername,
					Password: validPassword,
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
				userRegisterer: tt.fields.userRegisterer,
			}

			got, err := e.Execute(tt.args.ctx, tt.args.in)

			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)

				if tt.expectedErr != nil && err != nil {
					assert.EqualError(t, err, tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
