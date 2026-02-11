package post

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

func (h *handler) CreatePost(w http.ResponseWriter, r *http.Request) {
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

	var req CreatePostRequest
	if err := json.Read(r, &req); err != nil {
		log.Printf("failed to read post request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := h.service.CreatePost(r.Context(), int32(uid), req)
	if err != nil {
		log.Printf("failed to create post: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusCreated, post)
}

func (h *handler) GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	post, err := h.service.FindPostByID(r.Context(), int32(id))
	if err != nil {
		switch err {
		case ErrPostNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusOK, post)
}

func (h *handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
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

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	var req UpdatePostRequest
	if err := json.Read(r, &req); err != nil {
		log.Printf("failed to read update post request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := h.service.UpdatePost(r.Context(), int32(id), int32(uid), req)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			log.Printf("failed to update post: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusOK, post)
}

func (h *handler) DeletePost(w http.ResponseWriter, r *http.Request) {
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

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	err = h.service.DeletePost(r.Context(), int32(id), int32(uid))
	if err != nil {
		switch err {
		case ErrPostNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			log.Printf("failed to delete post: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) ListPostsByUserID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "user_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("failed to convert user_id to int: %v", err)
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	posts, err := h.service.ListPostsByUserID(r.Context(), int32(id))
	if err != nil {
		log.Printf("failed to list posts by user_id: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, posts)
}
