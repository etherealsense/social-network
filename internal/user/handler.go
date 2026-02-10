package user

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

func (h *handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		log.Printf("failed to list users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, users)
}

func (h *handler) GetMe(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	user, err := h.service.FindUserByID(r.Context(), int32(userIDFloat))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.Write(w, http.StatusOK, user)
}

func (h *handler) FindUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("failed to convert id to int: %v", err)
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.service.FindUserByID(r.Context(), int32(id))
	if err != nil {
		log.Printf("failed to find user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, user)
}

func (h *handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("failed to convert id to int: %v", err)
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.Read(r, &req); err != nil {
		log.Printf("failed to read user: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.UpdateUser(r.Context(), int32(id), req)
	if err != nil {
		log.Printf("failed to update user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, user)
}
