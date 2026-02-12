package chat

import (
	"context"
	"errors"
	"time"

	repo "github.com/etherealsense/social-network/internal/adapter/postgresql/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrChatAlreadyExists = errors.New("chat already exists between these users")
	ErrChatNotFound      = errors.New("chat not found")
	ErrSelfChat          = errors.New("cannot create chat with yourself")
)

type Service interface {
	CreateChat(ctx context.Context, userID int32, req CreateChatRequest) (repo.Chat, error)
	ListChatsByUserID(ctx context.Context, userID int32) ([]repo.Chat, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) CreateChat(ctx context.Context, userID int32, req CreateChatRequest) (repo.Chat, error) {
	if userID == req.UserID {
		return repo.Chat{}, ErrSelfChat
	}

	_, err := s.repo.GetChatByTwoUsers(ctx, repo.GetChatByTwoUsersParams{
		UserID:   userID,
		UserID_2: req.UserID,
	})
	if err == nil {
		return repo.Chat{}, ErrChatAlreadyExists
	}

	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}

	chat, err := s.repo.CreateChat(ctx, now)
	if err != nil {
		return repo.Chat{}, err
	}

	err = s.repo.CreateChatParticipant(ctx, repo.CreateChatParticipantParams{
		ChatID:   chat.ID,
		UserID:   userID,
		JoinedAt: now,
	})
	if err != nil {
		return repo.Chat{}, err
	}

	err = s.repo.CreateChatParticipant(ctx, repo.CreateChatParticipantParams{
		ChatID:   chat.ID,
		UserID:   req.UserID,
		JoinedAt: now,
	})
	if err != nil {
		return repo.Chat{}, err
	}

	return chat, nil
}

func (s *svc) ListChatsByUserID(ctx context.Context, userID int32) ([]repo.Chat, error) {
	return s.repo.ListChatsByUserID(ctx, userID)
}
