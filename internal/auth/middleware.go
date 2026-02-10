package auth

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

func Verifier(ja *JWTAuth) func(http.Handler) http.Handler {
	return jwtauth.Verifier(ja.GetTokenAuth())
}

func Authenticator(ja *JWTAuth) func(http.Handler) http.Handler {
	return jwtauth.Authenticator(ja.GetTokenAuth())
}
