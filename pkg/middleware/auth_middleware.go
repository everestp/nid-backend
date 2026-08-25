// pkg/middleware/auth_middleware.go

package middleware

import (
	"context"
	"net/http"
	"strings"

	"nid-backend/pkg/helpers"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var token string

		// Authorization header
		authHeader := strings.TrimSpace(
			r.Header.Get("Authorization"),
		)

		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimSpace(
				strings.TrimPrefix(authHeader, "Bearer "),
			)
		}

		// app_session cookie
		if token == "" {
			cookie, err := r.Cookie("app_session")

			if err == nil {
				token = strings.TrimSpace(cookie.Value)
			}
		}

		if token == "" {
			http.Error(
				w,
				"Unauthorized",
				http.StatusUnauthorized,
			)
			return
		}

		userID, valid := helpers.ValidateJWTToken(token)

		if !valid {
			http.Error(
				w,
				"Invalid token",
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			UserIDKey,
			userID,
		)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}
