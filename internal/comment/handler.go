package comment

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

func (h *handler) CreateComment(w http.ResponseWriter, r *http.Request) {
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

	var req CreateCommentRequest
	if err := json.Read(r, &req); err != nil {
		log.Printf("failed to read comment request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment, err := h.service.CreateComment(r.Context(), int32(postID), int32(uid), req)
	if err != nil {
		log.Printf("failed to create comment: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusCreated, comment)
}

func (h *handler) GetComment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	comment, err := h.service.FindCommentByID(r.Context(), int32(id))
	if err != nil {
		switch err {
		case ErrCommentNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusOK, comment)
}

func (h *handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	var req UpdateCommentRequest
	if err := json.Read(r, &req); err != nil {
		log.Printf("failed to read update comment request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	comment, err := h.service.UpdateComment(r.Context(), int32(id), int32(uid), req)
	if err != nil {
		switch err {
		case ErrCommentNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			log.Printf("failed to update comment: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusOK, comment)
}

func (h *handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteComment(r.Context(), int32(id), int32(uid))
	if err != nil {
		switch err {
		case ErrCommentNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			log.Printf("failed to delete comment: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) ListCommentsByPostID(w http.ResponseWriter, r *http.Request) {
	postIDStr := chi.URLParam(r, "post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post_id", http.StatusBadRequest)
		return
	}

	comments, err := h.service.ListCommentsByPostID(r.Context(), int32(postID))
	if err != nil {
		log.Printf("failed to list comments: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, comments)
}
