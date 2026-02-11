package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

func Verifier(ja *JWTAuth) func(http.Handler) http.Handler {
	return jwtauth.Verifier(ja.GetTokenAuth())
}

func Authenticator(ja *JWTAuth) func(http.Handler) http.Handler {
	return jwtauth.Authenticator(ja.GetTokenAuth())
}

func RequireAuth(ja *JWTAuth) func(chi.Router) {
	return func(r chi.Router) {
		r.Use(Verifier(ja))
		r.Use(Authenticator(ja))
		r.Use(ExtractUserID)
	}
}
