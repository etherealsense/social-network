package auth

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

type contextKey string

const userIDKey contextKey = "user_id"

func UserIDFromContext(ctx context.Context) int32 {
	userID, _ := ctx.Value(userIDKey).(int32)
	return userID
}

func ExtractUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		uid, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, int32(uid))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
