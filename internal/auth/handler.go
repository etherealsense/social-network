package auth

import (
	"net/http"

	"github.com/etherealsense/social-network/pkg/json"
)

const refreshTokenCookieName = "refresh_token"

type CookieConfig struct {
	Secure   bool
	SameSite http.SameSite
}

type Handler struct {
	service      Service
	jwtAuth      *JWTAuth
	cookieConfig CookieConfig
}

func NewHandler(service Service, jwtAuth *JWTAuth, cookieConfig CookieConfig) *Handler {
	return &Handler{
		service:      service,
		jwtAuth:      jwtAuth,
		cookieConfig: cookieConfig,
	}
}

func (h *Handler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieConfig.Secure,
		SameSite: h.cookieConfig.SameSite,
		MaxAge:   int(h.jwtAuth.refreshTokenTTL.Seconds()),
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.Read(r, &req); err != nil {
		http.Error(w, "failed to read register request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch err {
		case ErrUserAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	tokens, err := h.jwtAuth.GenerateTokenPair(int(user.ID))
	if err != nil {
		http.Error(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}

	h.setRefreshTokenCookie(w, tokens.RefreshToken)

	res := AuthResponse{
		AccessToken: tokens.AccessToken,
	}

	json.Write(w, http.StatusCreated, res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.Read(r, &req); err != nil {
		http.Error(w, "failed to read login request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(r.Context(), req)
	if err != nil {
		if err == ErrInvalidCredentials {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, "failed to login", http.StatusInternalServerError)
		}
		return
	}

	tokens, err := h.jwtAuth.GenerateTokenPair(int(user.ID))
	if err != nil {
		http.Error(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}

	h.setRefreshTokenCookie(w, tokens.RefreshToken)

	res := AuthResponse{
		AccessToken: tokens.AccessToken,
	}

	json.Write(w, http.StatusOK, res)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		http.Error(w, "refresh token not found", http.StatusUnauthorized)
		return
	}

	userID, err := h.jwtAuth.ValidateRefreshToken(cookie.Value)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	tokens, err := h.jwtAuth.GenerateTokenPair(userID)
	if err != nil {
		http.Error(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}

	h.setRefreshTokenCookie(w, tokens.RefreshToken)

	res := AuthResponse{
		AccessToken: tokens.AccessToken,
	}

	json.Write(w, http.StatusOK, res)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieConfig.Secure,
		SameSite: h.cookieConfig.SameSite,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}
