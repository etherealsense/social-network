package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func RequireAuth(h *Handler) func(chi.Router) {
	return func(r chi.Router) {
		r.Use(jwtauth.Verifier(h.jwtAuth.GetTokenAuth()))
		r.Use(jwtauth.Authenticator(h.jwtAuth.GetTokenAuth()))
		r.Use(ExtractUserID)
	}
}
