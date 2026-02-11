package follow

import (
	"log"
	"net/http"
	"strconv"

	"github.com/etherealsense/social-network/pkg/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) FollowUser(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uid, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	followingIDStr := chi.URLParam(r, "user_id")
	followingID, err := strconv.Atoi(followingIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	follow, err := h.service.FollowUser(r.Context(), int32(uid), int32(followingID))
	if err != nil {
		switch err {
		case ErrSelfFollow:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case ErrAlreadyFollowing:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			log.Printf("failed to follow user: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusCreated, follow)
}

func (h *handler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uid, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	followingIDStr := chi.URLParam(r, "user_id")
	followingID, err := strconv.Atoi(followingIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	err = h.service.UnfollowUser(r.Context(), int32(uid), int32(followingID))
	if err != nil {
		switch err {
		case ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			log.Printf("failed to unfollow user: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	followers, err := h.service.ListFollowers(r.Context(), int32(userID))
	if err != nil {
		log.Printf("failed to list followers: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	following, err := h.service.ListFollowing(r.Context(), int32(userID))
	if err != nil {
		log.Printf("failed to list following: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, following)
}
