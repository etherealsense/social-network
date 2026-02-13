package chat

import "github.com/jackc/pgx/v5/pgtype"

type CreateChatRequest struct {
	UserID int32 `json:"user_id"`
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

type MessageResponse struct {
	ID        int32              `json:"id"`
	ChatID    int32              `json:"chat_id"`
	SenderID  int32              `json:"sender_id"`
	Content   string             `json:"content"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}
