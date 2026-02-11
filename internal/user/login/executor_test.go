package login

import (
	"context"
	"errors"
	"testing"
	"time"

	"mkk-luna-test-task/internal/user"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type stubUserGetterSuccess struct {
	returnUser *user.Model
}

func (s *stubUserGetterSuccess) GetUser(ctx context.Context, username string) (*user.Model, error) {
	return s.returnUser, nil
}

type stubUserGetterFail struct {
	err error
}

func (s *stubUserGetterFail) GetUser(ctx context.Context, username string) (*user.Model, error) {
	return nil, s.err
}

func hashedPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func TestExecutor_Execute(t *testing.T) {
	validUsername := "user1"
	validPassword := "password123"
	validHashedPassword := hashedPassword(validPassword)
	validUser := &user.Model{
		Id:             5,
		Username:       validUsername,
		PasswordHashed: validHashedPassword,
	}

	type fields struct {
		userGetter userGetter
		jwtSecret  []byte
		jwtExpiry  time.Duration
	}
	type args struct {
		ctx context.Context
		in  LoginInput
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *LoginResult
		wantErr     bool
		expectedErr error
		checkToken  bool
	}{
		{
			name: "success login",
			fields: fields{
				userGetter: &stubUserGetterSuccess{
					returnUser: validUser,
				},
				jwtSecret: []byte("secret"),
				jwtExpiry: time.Hour,
			},
			args: args{
				ctx: context.Background(),
				in: LoginInput{
					Username: validUsername,
					Password: validPassword,
				},
			},
			want:       &LoginResult{},
			wantErr:    false,
			checkToken: true,
		},
		{
			name: "missing username",
			fields: fields{
				userGetter: &stubUserGetterSuccess{validUser},
				jwtSecret:  []byte("secret"),
				jwtExpiry:  time.Hour,
			},
			args: args{
				ctx: context.Background(),
				in: LoginInput{
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
				userGetter: &stubUserGetterSuccess{validUser},
				jwtSecret:  []byte("secret"),
				jwtExpiry:  time.Hour,
			},
			args: args{
				ctx: context.Background(),
				in: LoginInput{
					Username: validUsername,
					Password: "",
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("username and password required"),
		},
		{
			name: "user getter returns error",
			fields: fields{
				userGetter: &stubUserGetterFail{err: errors.New("db error")},
				jwtSecret:  []byte("secret"),
				jwtExpiry:  time.Hour,
			},
			args: args{
				ctx: context.Background(),
				in: LoginInput{
					Username: validUsername,
					Password: validPassword,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("invalid username or password"),
		},
		{
			name: "user not found",
			fields: fields{
				userGetter: &stubUserGetterSuccess{returnUser: nil},
				jwtSecret:  []byte("secret"),
				jwtExpiry:  time.Hour,
			},
			args: args{
				ctx: context.Background(),
				in: LoginInput{
					Username: validUsername,
					Password: validPassword,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("invalid username or password"),
		},
		{
			name: "wrong password",
			fields: fields{
				userGetter: &stubUserGetterSuccess{
					returnUser: &user.Model{
						Id:             7,
						Username:       validUsername,
						PasswordHashed: hashedPassword("otherpass"),
					},
				},
				jwtSecret: []byte("secret"),
				jwtExpiry: time.Hour,
			},
			args: args{
				ctx: context.Background(),
				in: LoginInput{
					Username: validUsername,
					Password: validPassword,
				},
			},
			want:        nil,
			wantErr:     true,
			expectedErr: errors.New("invalid username or password"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &executor{
				userGetter: tt.fields.userGetter,
				jwtSecret:  tt.fields.jwtSecret,
				jwtExpiry:  tt.fields.jwtExpiry,
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

				if tt.checkToken && got != nil {
					assert.NotEmpty(t, got.Token)

					parsed, perr := jwt.Parse(got.Token, func(token *jwt.Token) (interface{}, error) {
						return tt.fields.jwtSecret, nil
					})

					assert.NoError(t, perr)

					if claims, ok := parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
						assert.Equal(t, float64(validUser.Id), claims["id"])
						assert.Equal(t, validUsername, claims["name"])
						assert.Contains(t, claims, "exp")
					} else {
						t.Errorf("token invalid or no claims: %v", got.Token)
					}
				}
			}
		})
	}
}
