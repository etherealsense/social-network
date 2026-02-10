package user

import (
	"context"
	"errors"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
	"github.com/etherealsense/social-network/pkg/crypto"
	"github.com/etherealsense/social-network/pkg/validator"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type Service interface {
	ListUsers(ctx context.Context) ([]repo.User, error)
	FindUserByID(ctx context.Context, id int32) (UserResponse, error)
	UpdateUser(ctx context.Context, id int32, req UpdateUserRequest) (repo.UpdateUserRow, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) ListUsers(ctx context.Context) ([]repo.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *svc) FindUserByID(ctx context.Context, id int32) (UserResponse, error) {
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return UserResponse{}, ErrUserNotFound
	}

	return UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *svc) CreateUser(ctx context.Context, user repo.CreateUserParams) (repo.CreateUserRow, error) {
	err := validator.ValidateEmail(user.Email)
	if err != nil {
		return repo.CreateUserRow{}, err
	}

	_, err = s.repo.FindUserByEmail(ctx, user.Email)
	if err == nil {
		return repo.CreateUserRow{}, ErrUserAlreadyExists
	}

	err = validator.ValidatePassword(user.Password)
	if err != nil {
		return repo.CreateUserRow{}, err
	}

	hashedPassword, err := crypto.HashPassword(user.Password)
	if err != nil {
		return repo.CreateUserRow{}, err
	}
	user.Password = hashedPassword

	return s.repo.CreateUser(ctx, user)
}

func (s *svc) UpdateUser(ctx context.Context, id int32, req UpdateUserRequest) (repo.UpdateUserRow, error) {
	_, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return repo.UpdateUserRow{}, ErrUserNotFound
	}

	params := repo.UpdateUserParams{
		ID: id,
	}

	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}

	if req.Email != nil {
		err = validator.ValidateEmail(*req.Email)
		if err != nil {
			return repo.UpdateUserRow{}, err
		}
		params.Email = pgtype.Text{String: *req.Email, Valid: true}
	}

	if req.Password != nil {
		err = validator.ValidatePassword(*req.Password)
		if err != nil {
			return repo.UpdateUserRow{}, err
		}

		hashedPassword, err := crypto.HashPassword(*req.Password)
		if err != nil {
			return repo.UpdateUserRow{}, err
		}
		params.Password = pgtype.Text{String: hashedPassword, Valid: true}
	}

	return s.repo.UpdateUser(ctx, params)
}
