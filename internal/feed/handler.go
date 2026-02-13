package feed

import (
	"log/slog"
	"net/http"

	"github.com/etherealsense/social-network/internal/auth"
	"github.com/etherealsense/social-network/pkg/json"
	"github.com/etherealsense/social-network/pkg/pagination"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetFeed(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserIDFromContext(r.Context())
	p := pagination.Parse(r)

	posts, err := h.service.GetFeed(r.Context(), uid, p.Limit, p.Offset)
	if err != nil {
		slog.Error("failed to get feed", "error", err, "user_id", uid)
		http.Error(w, "failed to get feed", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, posts)
}
