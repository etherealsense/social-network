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
	ErrUserNotFound      = errors.New("user not found")
	ErrNotParticipant    = errors.New("user is not a participant of this chat")
)

type Service interface {
	CreateChat(ctx context.Context, userID int32, req CreateChatRequest) (repo.Chat, error)
	ListChatsByUserID(ctx context.Context, userID, limit, offset int32) ([]repo.Chat, error)
	ListParticipantsByChatID(ctx context.Context, chatID, limit, offset int32) ([]repo.ChatParticipant, error)
	CreateMessage(ctx context.Context, chatID, senderID int32, content string) (repo.Message, error)
	ListMessagesByChatID(ctx context.Context, chatID, limit, offset int32) ([]repo.Message, error)
	IsParticipant(ctx context.Context, chatID, userID int32) error
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

	_, err := s.repo.FindUserByID(ctx, req.UserID)
	if err != nil {
		return repo.Chat{}, ErrUserNotFound
	}

	_, err = s.repo.GetChatByTwoUsers(ctx, repo.GetChatByTwoUsersParams{
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

func (s *svc) ListChatsByUserID(ctx context.Context, userID, limit, offset int32) ([]repo.Chat, error) {
	return s.repo.ListChatsByUserID(ctx, repo.ListChatsByUserIDParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *svc) ListParticipantsByChatID(ctx context.Context, chatID, limit, offset int32) ([]repo.ChatParticipant, error) {
	return s.repo.ListChatParticipantsByChatID(ctx, repo.ListChatParticipantsByChatIDParams{
		ChatID: chatID,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *svc) CreateMessage(ctx context.Context, chatID, senderID int32, content string) (repo.Message, error) {
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}

	return s.repo.CreateMessage(ctx, repo.CreateMessageParams{
		ChatID:    chatID,
		SenderID:  senderID,
		Content:   content,
		CreatedAt: now,
		IsRead:    false,
	})
}

func (s *svc) ListMessagesByChatID(ctx context.Context, chatID, limit, offset int32) ([]repo.Message, error) {
	return s.repo.ListMessagesByChatID(ctx, repo.ListMessagesByChatIDParams{
		ChatID: chatID,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *svc) IsParticipant(ctx context.Context, chatID, userID int32) error {
	_, err := s.repo.GetChatParticipantByChatIDAndUserID(ctx, repo.GetChatParticipantByChatIDAndUserIDParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return ErrNotParticipant
	}
	return nil
}
