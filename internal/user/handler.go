package user

import (
	"log"
	"net/http"

	"github.com/etherealsense/social-network/internal/auth"
	"github.com/etherealsense/social-network/pkg/json"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{service: service}
}

func (h *handler) GetMe(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())

	user, err := h.service.FindUserByID(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.Write(w, http.StatusOK, user)
}

func (h *handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req UpdateUserRequest
	if err := json.Read(r, &req); err != nil {
		log.Printf("failed to read user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.UpdateUser(r.Context(), userID, req)
	if err != nil {
		log.Printf("failed to update user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, user)
}
