package register

import (
	"context"
	"fmt"

	"mkk-luna-test-task/internal/user"

	"golang.org/x/crypto/bcrypt"
)

type userRegisterer interface {
	RegisterUser(ctx context.Context, username, passwordHashed string) (*user.Model, error)
}

type executor struct {
	userRegisterer userRegisterer
}

func NewExecutor(userRegisterer userRegisterer) *executor {
	return &executor{userRegisterer: userRegisterer}
}

type RegisterInput struct {
	Username string
	Password string
}

type RegisterResult struct {
	UserId int
}

func (e *executor) Execute(ctx context.Context, in RegisterInput) (*RegisterResult, error) {
	if in.Username == "" || in.Password == "" {
		return nil, fmt.Errorf("username and password required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := e.userRegisterer.RegisterUser(ctx, in.Username, string(hashedPassword))

	if err != nil {
		return nil, err
	}

	return &RegisterResult{UserId: user.Id}, nil
}
