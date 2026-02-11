package login

import (
	"context"
	"fmt"
	"time"

	"mkk-luna-test-task/internal/user"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type userGetter interface {
	GetUser(ctx context.Context, username string) (*user.Model, error)
}

type executor struct {
	userGetter userGetter
	jwtSecret  []byte
	jwtExpiry  time.Duration
}

func NewExecutor(userGetter userGetter, jwtSecret []byte, jwtExpiry time.Duration) *executor {
	return &executor{
		userGetter: userGetter,
		jwtSecret:  jwtSecret,
		jwtExpiry:  jwtExpiry,
	}
}

type LoginInput struct {
	Username string
	Password string
}

type LoginResult struct {
	Token string
}

func (e *executor) Execute(ctx context.Context, in LoginInput) (*LoginResult, error) {
	if in.Username == "" || in.Password == "" {
		return nil, fmt.Errorf("username and password required")
	}

	u, err := e.userGetter.GetUser(ctx, in.Username)

	if err != nil || u == nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHashed), []byte(in.Password))

	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	claims := jwt.MapClaims{
		"id":   u.Id,
		"name": u.Username,
		"exp":  time.Now().Add(e.jwtExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(e.jwtSecret)

	if err != nil {
		return nil, fmt.Errorf("could not create token: %w", err)
	}

	return &LoginResult{Token: tokenString}, nil
}
