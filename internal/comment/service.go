package comment

import (
	"context"
	"errors"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCommentForbidden = errors.New("forbidden")
)

type Service interface {
	CreateComment(ctx context.Context, postID, userID int32, req CreateCommentRequest) (repo.Comment, error)
	FindCommentByID(ctx context.Context, id int32) (repo.Comment, error)
	UpdateComment(ctx context.Context, id int32, userID int32, req UpdateCommentRequest) (repo.Comment, error)
	DeleteComment(ctx context.Context, id int32, userID int32) error
	ListCommentsByPostID(ctx context.Context, postID int32) ([]repo.Comment, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) CreateComment(ctx context.Context, postID, userID int32, req CreateCommentRequest) (repo.Comment, error) {
	return s.repo.CreateComment(ctx, repo.CreateCommentParams{
		PostID:  postID,
		UserID:  userID,
		Content: req.Content,
	})
}

func (s *svc) FindCommentByID(ctx context.Context, id int32) (repo.Comment, error) {
	c, err := s.repo.FindCommentByID(ctx, id)
	if err != nil {
		return repo.Comment{}, ErrCommentNotFound
	}
	return c, nil
}

func (s *svc) UpdateComment(ctx context.Context, id int32, userID int32, req UpdateCommentRequest) (repo.Comment, error) {
	c, err := s.repo.FindCommentByID(ctx, id)
	if err != nil {
		return repo.Comment{}, ErrCommentNotFound
	}

	if c.UserID != userID {
		return repo.Comment{}, ErrCommentForbidden
	}

	params := repo.UpdateCommentParams{
		ID: id,
	}

	if req.Content != nil {
		params.Content = pgtype.Text{String: *req.Content, Valid: true}
	}

	return s.repo.UpdateComment(ctx, params)
}

func (s *svc) DeleteComment(ctx context.Context, id int32, userID int32) error {
	c, err := s.repo.FindCommentByID(ctx, id)
	if err != nil {
		return ErrCommentNotFound
	}

	if c.UserID != userID {
		return ErrCommentForbidden
	}

	return s.repo.DeleteComment(ctx, id)
}

func (s *svc) ListCommentsByPostID(ctx context.Context, postID int32) ([]repo.Comment, error) {
	return s.repo.ListCommentsByPostID(ctx, postID)
}
