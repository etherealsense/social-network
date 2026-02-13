package auth

import (
	"context"
	"errors"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
	"github.com/etherealsense/social-network/pkg/crypto"
	"github.com/etherealsense/social-network/pkg/database"
	"github.com/etherealsense/social-network/pkg/validator"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (repo.CreateUserRow, error)
	Login(ctx context.Context, req LoginRequest) (repo.CreateUserRow, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) Register(ctx context.Context, req RegisterRequest) (repo.CreateUserRow, error) {
	err := validator.ValidateEmail(req.Email)
	if err != nil {
		return repo.CreateUserRow{}, err
	}

	err = validator.ValidatePassword(req.Password)
	if err != nil {
		return repo.CreateUserRow{}, err
	}

	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return repo.CreateUserRow{}, err
	}

	user, err := s.repo.CreateUser(ctx, repo.CreateUserParams{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	})
	if err != nil {
		if database.IsUniqueViolation(err) {
			return repo.CreateUserRow{}, ErrUserAlreadyExists
		}

		return repo.CreateUserRow{}, err
	}

	return user, nil
}

func (s *svc) Login(ctx context.Context, req LoginRequest) (repo.CreateUserRow, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return repo.CreateUserRow{}, ErrInvalidCredentials
	}

	err = crypto.ComparePassword(user.Password, req.Password)
	if err != nil {
		return repo.CreateUserRow{}, ErrInvalidCredentials
	}

	return repo.CreateUserRow{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
