package chat

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/coder/websocket"
	"github.com/etherealsense/social-network/internal/auth"
	jsonpkg "github.com/etherealsense/social-network/pkg/json"
	"github.com/etherealsense/social-network/pkg/pagination"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
	hub     *Hub
}

func NewHandler(service Service, hub *Hub) *Handler {
	return &Handler{service: service, hub: hub}
}

func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	var req CreateChatRequest
	if err := jsonpkg.Read(r, &req); err != nil {
		slog.Error("failed to read chat request", "error", err)
		http.Error(w, "failed to read chat request", http.StatusBadRequest)
		return
	}

	chat, err := h.service.CreateChat(r.Context(), uid, req)
	if err != nil {
		switch err {
		case ErrSelfChat:
			http.Error(w, "cannot create chat with yourself", http.StatusBadRequest)
		case ErrChatAlreadyExists:
			http.Error(w, "chat already exists between these users", http.StatusConflict)
		case ErrUserNotFound:
			http.Error(w, "user not found", http.StatusNotFound)
		default:
			slog.Error("failed to create chat", "error", err, "user_id", uid)
			http.Error(w, "failed to create chat", http.StatusInternalServerError)
		}
		return
	}

	jsonpkg.Write(w, http.StatusCreated, chat)
}

func (h *Handler) ListChats(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())
	p := pagination.Parse(r)

	chats, err := h.service.ListChatsByUserID(r.Context(), uid, p.Limit, p.Offset)
	if err != nil {
		slog.Error("failed to list chats", "error", err, "user_id", uid)
		http.Error(w, "failed to list chats", http.StatusInternalServerError)
		return
	}

	jsonpkg.Write(w, http.StatusOK, chats)
}

func (h *Handler) ListParticipants(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "chat_id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	p := pagination.Parse(r)

	participants, err := h.service.ListParticipantsByChatID(r.Context(), int32(chatID), p.Limit, p.Offset)
	if err != nil {
		slog.Error("failed to list participants", "error", err, "chat_id", chatID)
		http.Error(w, "failed to list participants", http.StatusInternalServerError)
		return
	}

	jsonpkg.Write(w, http.StatusOK, participants)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	chatIDStr := chi.URLParam(r, "chat_id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	if err := h.service.IsParticipant(r.Context(), int32(chatID), uid); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	p := pagination.Parse(r)

	messages, err := h.service.ListMessagesByChatID(r.Context(), int32(chatID), p.Limit, p.Offset)
	if err != nil {
		slog.Error("failed to list messages", "error", err, "chat_id", chatID)
		http.Error(w, "failed to list messages", http.StatusInternalServerError)
		return
	}

	jsonpkg.Write(w, http.StatusOK, messages)
}

// HandleWebSocket upgrades the connection and allows sending/receiving messages in real time.
// Clients connect to GET /api/v1/chats/{chat_id}/ws and send JSON: {"content": "hello"}
// The server broadcasts MessageResponse JSON to all connected participants.
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	chatIDStr := chi.URLParam(r, "chat_id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	if err := h.service.IsParticipant(r.Context(), int32(chatID), uid); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		slog.Error("failed to accept websocket", "error", err)
		return
	}
	defer conn.CloseNow()

	c := &client{
		conn:   conn,
		userID: uid,
		chatID: int32(chatID),
	}

	h.hub.Register(c)
	defer h.hub.Unregister(c)

	for {
		_, data, err := conn.Read(r.Context())
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				break
			}
			slog.Error("failed to read websocket message", "error", err, "user_id", uid)
			break
		}

		var req SendMessageRequest
		if err := json.Unmarshal(data, &req); err != nil {
			slog.Error("invalid websocket message payload", "error", err)
			continue
		}

		if req.Content == "" {
			continue
		}

		msg, err := h.service.CreateMessage(r.Context(), int32(chatID), uid, req.Content)
		if err != nil {
			slog.Error("failed to create message", "error", err, "chat_id", chatID)
			continue
		}

		h.hub.Broadcast(int32(chatID), MessageResponse{
			ID:        msg.ID,
			ChatID:    msg.ChatID,
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
		})
	}

	conn.Close(websocket.StatusNormalClosure, "")
}
