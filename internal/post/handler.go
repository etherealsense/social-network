package post

import (
	"log"
	"net/http"
	"strconv"

	"github.com/etherealsense/social-network/pkg/json"
	"github.com/go-chi/chi/v5"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
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
