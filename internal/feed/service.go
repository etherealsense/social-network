package feed

import (
	"context"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
)

type Service interface {
	GetFeed(ctx context.Context, userID int32, limit, offset int32) ([]repo.GetFeedRow, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) GetFeed(ctx context.Context, userID int32, limit, offset int32) ([]repo.GetFeedRow, error) {
	return s.repo.GetFeed(ctx, repo.GetFeedParams{
		FollowerID: userID,
		Limit:      limit,
		Offset:     offset,
	})
}
