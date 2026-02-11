package like

import (
	"context"
	"errors"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
)

var (
	ErrAlreadyLiked = errors.New("already liked this post")
	ErrPostNotFound = errors.New("post not found")
)

type Service interface {
	LikePost(ctx context.Context, userID, postID int32) (repo.Like, error)
	UnlikePost(ctx context.Context, userID, postID int32) error
	ListLikesByPostID(ctx context.Context, postID int32, limit, offset int32) ([]repo.Like, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) LikePost(ctx context.Context, userID, postID int32) (repo.Like, error) {
	_, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		return repo.Like{}, ErrPostNotFound
	}

	l, err := s.repo.LikePost(ctx, repo.LikePostParams{
		UserID: userID,
		PostID: postID,
	})
	if err != nil {
		return repo.Like{}, ErrAlreadyLiked
	}
	return l, nil
}

func (s *svc) UnlikePost(ctx context.Context, userID, postID int32) error {
	_, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	return s.repo.UnlikePost(ctx, repo.UnlikePostParams{
		UserID: userID,
		PostID: postID,
	})
}

func (s *svc) ListLikesByPostID(ctx context.Context, postID int32, limit, offset int32) ([]repo.Like, error) {
	return s.repo.ListLikesByPostID(ctx, repo.ListLikesByPostIDParams{
		PostID: postID,
		Limit:  limit,
		Offset: offset,
	})
}
