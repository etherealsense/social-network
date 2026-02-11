package like

import (
	"log"
	"net/http"
	"strconv"

	"github.com/etherealsense/social-network/pkg/json"
	"github.com/etherealsense/social-network/pkg/pagination"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) LikePost(w http.ResponseWriter, r *http.Request) {
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

	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post_id", http.StatusBadRequest)
		return
	}

	l, err := h.service.LikePost(r.Context(), int32(uid), int32(postID))
	if err != nil {
		switch err {
		case ErrPostNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrAlreadyLiked:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			log.Printf("failed to like post: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusCreated, l)
}

func (h *handler) UnlikePost(w http.ResponseWriter, r *http.Request) {
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

	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post_id", http.StatusBadRequest)
		return
	}

	err = h.service.UnlikePost(r.Context(), int32(uid), int32(postID))
	if err != nil {
		switch err {
		case ErrPostNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			log.Printf("failed to unlike post: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) ListLikesByPostID(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post_id", http.StatusBadRequest)
		return
	}

	p := pagination.Parse(r)

	likes, err := h.service.ListLikesByPostID(r.Context(), int32(postID), p.Limit, p.Offset)
	if err != nil {
		log.Printf("failed to list likes: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, likes)
}
