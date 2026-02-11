package user

import (
	"log/slog"
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
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	json.Write(w, http.StatusOK, user)
}

func (h *handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req UpdateUserRequest
	if err := json.Read(r, &req); err != nil {
		slog.Error("failed to read user", "error", err)
		http.Error(w, "failed to read user", http.StatusBadRequest)
		return
	}

	user, err := h.service.UpdateUser(r.Context(), userID, req)
	if err != nil {
		slog.Error("failed to update user", "error", err, "user_id", userID)
		http.Error(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, user)
}
