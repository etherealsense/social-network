package post

import (
	"log"
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

func (h *handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	var req CreatePostRequest
	if err := json.Read(r, &req); err != nil {
		log.Printf("failed to read post request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := h.service.CreatePost(r.Context(), uid, req)
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
	uid := auth.UserIDFromContext(r.Context())

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

	post, err := h.service.UpdatePost(r.Context(), int32(id), uid, req)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrPostForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			log.Printf("failed to update post: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusOK, post)
}

func (h *handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	err = h.service.DeletePost(r.Context(), int32(id), uid)
	if err != nil {
		switch err {
		case ErrPostNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrPostForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
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

	p := pagination.Parse(r)

	posts, err := h.service.ListPostsByUserID(r.Context(), int32(id), p.Limit, p.Offset)
	if err != nil {
		log.Printf("failed to list posts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, posts)
}
