package like

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/etherealsense/social-network/internal/auth"
	"github.com/etherealsense/social-network/pkg/json"
	"github.com/etherealsense/social-network/pkg/pagination"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) LikePost(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post_id", http.StatusBadRequest)
		return
	}

	l, err := h.service.LikePost(r.Context(), uid, int32(postID))
	if err != nil {
		switch err {
		case ErrPostNotFound:
			http.Error(w, "post not found", http.StatusNotFound)
		case ErrAlreadyLiked:
			http.Error(w, "already liked this post", http.StatusConflict)
		default:
			slog.Error("failed to like post", "error", err)
			http.Error(w, "failed to like post", http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusCreated, l)
}

func (h *Handler) UnlikePost(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post_id", http.StatusBadRequest)
		return
	}

	err = h.service.UnlikePost(r.Context(), uid, int32(postID))
	if err != nil {
		switch err {
		case ErrPostNotFound:
			http.Error(w, "post not found", http.StatusNotFound)
		default:
			slog.Error("failed to unlike post", "error", err)
			http.Error(w, "failed to unlike post", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListLikesByPostID(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post_id", http.StatusBadRequest)
		return
	}

	p := pagination.Parse(r)

	likes, err := h.service.ListLikesByPostID(r.Context(), int32(postID), p.Limit, p.Offset)
	if err != nil {
		slog.Error("failed to list likes", "error", err, "post_id", postID)
		http.Error(w, "failed to list likes", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, likes)
}
