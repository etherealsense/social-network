package chat

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/etherealsense/social-network/internal/auth"
	"github.com/etherealsense/social-network/pkg/json"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	var req CreateChatRequest
	if err := json.Read(r, &req); err != nil {
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

	json.Write(w, http.StatusCreated, chat)
}

func (h *Handler) ListChats(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	chats, err := h.service.ListChatsByUserID(r.Context(), uid)
	if err != nil {
		slog.Error("failed to list chats", "error", err, "user_id", uid)
		http.Error(w, "failed to list chats", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, chats)
}

func (h *Handler) ListParticipants(w http.ResponseWriter, r *http.Request) {
	chatIDStr := chi.URLParam(r, "chat_id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	participants, err := h.service.ListParticipantsByChatID(r.Context(), int32(chatID))
	if err != nil {
		slog.Error("failed to list participants", "error", err, "chat_id", chatID)
		http.Error(w, "failed to list participants", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, participants)
}
