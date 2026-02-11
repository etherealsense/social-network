package follow

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/etherealsense/social-network/internal/auth"
	"github.com/etherealsense/social-network/pkg/json"
	"github.com/etherealsense/social-network/pkg/pagination"
	"github.com/go-chi/chi/v5"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) FollowUser(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	followingIDStr := chi.URLParam(r, "user_id")
	followingID, err := strconv.Atoi(followingIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	follow, err := h.service.FollowUser(r.Context(), uid, int32(followingID))
	if err != nil {
		switch err {
		case ErrSelfFollow:
			http.Error(w, "cannot follow yourself", http.StatusBadRequest)
		case ErrAlreadyFollowing:
			http.Error(w, "already following this user", http.StatusConflict)
		default:
			slog.Error("failed to follow user", "error", err)
			http.Error(w, "failed to follow user", http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusCreated, follow)
}

func (h *handler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	followingIDStr := chi.URLParam(r, "user_id")
	followingID, err := strconv.Atoi(followingIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	err = h.service.UnfollowUser(r.Context(), uid, int32(followingID))
	if err != nil {
		switch err {
		case ErrUserNotFound:
			http.Error(w, "user not found", http.StatusNotFound)
		default:
			slog.Error("failed to unfollow user", "error", err)
			http.Error(w, "failed to unfollow user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) ListFollowers(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	p := pagination.Parse(r)

	followers, err := h.service.ListFollowers(r.Context(), int32(userID), p.Limit, p.Offset)
	if err != nil {
		slog.Error("failed to list followers", "error", err, "user_id", userID)
		http.Error(w, "failed to list followers", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, followers)
}

func (h *handler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	p := pagination.Parse(r)

	following, err := h.service.ListFollowing(r.Context(), int32(userID), p.Limit, p.Offset)
	if err != nil {
		slog.Error("failed to list following", "error", err, "user_id", userID)
		http.Error(w, "failed to list following", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, following)
}
