package follow

import (
	"context"
	"errors"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
)

var (
	ErrAlreadyFollowing = errors.New("already following this user")
	ErrSelfFollow       = errors.New("cannot follow yourself")
	ErrUserNotFound     = errors.New("user not found")
)

type Service interface {
	FollowUser(ctx context.Context, followerID, followingID int32) (repo.Follow, error)
	UnfollowUser(ctx context.Context, followerID, followingID int32) error
	ListFollowers(ctx context.Context, userID int32, limit, offset int32) ([]repo.Follow, error)
	ListFollowing(ctx context.Context, userID int32, limit, offset int32) ([]repo.Follow, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) FollowUser(ctx context.Context, followerID, followingID int32) (repo.Follow, error) {
	if followerID == followingID {
		return repo.Follow{}, ErrSelfFollow
	}

	f, err := s.repo.FollowUser(ctx, repo.FollowUserParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
	if err != nil {
		return repo.Follow{}, ErrAlreadyFollowing
	}
	return f, nil
}

func (s *svc) UnfollowUser(ctx context.Context, followerID, followingID int32) error {
	_, err := s.repo.FindUserByID(ctx, followingID)
	if err != nil {
		return ErrUserNotFound
	}

	return s.repo.UnfollowUser(ctx, repo.UnfollowUserParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
}

func (s *svc) ListFollowers(ctx context.Context, userID int32, limit, offset int32) ([]repo.Follow, error) {
	return s.repo.ListFollowers(ctx, repo.ListFollowersParams{
		FollowingID: userID,
		Limit:       limit,
		Offset:      offset,
	})
}

func (s *svc) ListFollowing(ctx context.Context, userID int32, limit, offset int32) ([]repo.Follow, error) {
	return s.repo.ListFollowing(ctx, repo.ListFollowingParams{
		FollowerID: userID,
		Limit:      limit,
		Offset:     offset,
	})
}
