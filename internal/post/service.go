package post

import (
	"context"
	"errors"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrPostAlreadyExists = errors.New("post already exists")
	ErrPostNotFound      = errors.New("post not found")
)

type Service interface {
	CreatePost(ctx context.Context, userID int32, req CreatePostRequest) (repo.Post, error)
	FindPostByID(ctx context.Context, id int32) (repo.Post, error)
	UpdatePost(ctx context.Context, id int32, userID int32, req UpdatePostRequest) (repo.Post, error)
	DeletePost(ctx context.Context, id int32, userID int32) error
	ListPostsByUserID(ctx context.Context, userID int32) ([]repo.Post, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) CreatePost(ctx context.Context, userID int32, req CreatePostRequest) (repo.Post, error) {
	return s.repo.CreatePost(ctx, repo.CreatePostParams{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	})
}

func (s *svc) FindPostByID(ctx context.Context, id int32) (repo.Post, error) {
	post, err := s.repo.FindPostByID(ctx, id)
	if err != nil {
		return repo.Post{}, ErrPostNotFound
	}
	return post, nil
}

func (s *svc) UpdatePost(ctx context.Context, id int32, userID int32, req UpdatePostRequest) (repo.Post, error) {
	post, err := s.repo.FindPostByID(ctx, id)
	if err != nil {
		return repo.Post{}, ErrPostNotFound
	}

	if post.UserID != userID {
		return repo.Post{}, errors.New("forbidden")
	}

	params := repo.UpdatePostParams{
		ID: id,
	}

	if req.Title != nil {
		params.Title = pgtype.Text{String: *req.Title, Valid: true}
	}

	if req.Content != nil {
		params.Content = pgtype.Text{String: *req.Content, Valid: true}
	}

	return s.repo.UpdatePost(ctx, params)
}

func (s *svc) DeletePost(ctx context.Context, id int32, userID int32) error {
	post, err := s.repo.FindPostByID(ctx, id)
	if err != nil {
		return ErrPostNotFound
	}

	if post.UserID != userID {
		return errors.New("forbidden")
	}

	return s.repo.DeletePost(ctx, id)
}

func (s *svc) ListPostsByUserID(ctx context.Context, userID int32) ([]repo.Post, error) {
	return s.repo.ListPostsByUserID(ctx, userID)
}
