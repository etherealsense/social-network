package comment

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

func (h *handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

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

	comment, err := h.service.CreateComment(r.Context(), int32(postID), uid, req)
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
	uid := auth.UserIDFromContext(r.Context())

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

	comment, err := h.service.UpdateComment(r.Context(), int32(id), uid, req)
	if err != nil {
		switch err {
		case ErrCommentNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrCommentForbidden:
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
	uid := auth.UserIDFromContext(r.Context())

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteComment(r.Context(), int32(id), uid)
	if err != nil {
		switch err {
		case ErrCommentNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrCommentForbidden:
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

	p := pagination.Parse(r)

	comments, err := h.service.ListCommentsByPostID(r.Context(), int32(postID), p.Limit, p.Offset)
	if err != nil {
		log.Printf("failed to list comments: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, comments)
}
