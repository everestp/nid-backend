// pkg/middleware/auth_middleware.go

package middleware

import (
	"context"
	"net/http"
	"strings"

	"nid-backend/pkg/helpers"
)

// ============================================================
// Context
// ============================================================

type contextKey string

const UserIDKey contextKey = "user_id"

// ============================================================
// Internal NID Authentication Middleware
// ============================================================
//
// This middleware is ONLY for NID's own protected APIs.
//
// Supported:
//
//   Authorization: Bearer <internal-session-token>
//
// OR:
//
//   Cookie: nid_token=<internal-session-token>
//
// It does NOT validate:
//
//   - OIDC access tokens
//   - OIDC ID tokens
//   - OAuth authorization codes
//
// ============================================================

func AuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {

		var token string

		// ========================================================
		// 1. Authorization Header
		// ========================================================

		authHeader := strings.TrimSpace(
			r.Header.Get("Authorization"),
		)

		if strings.HasPrefix(
			authHeader,
			"Bearer ",
		) {
			token = strings.TrimSpace(
				strings.TrimPrefix(
					authHeader,
					"Bearer ",
				),
			)
		}

		// ========================================================
		// 2. NID Internal Session Cookie
		// ========================================================

		if token == "" {

			cookie, err := r.Cookie(
				"nid_token",
			)

			if err == nil {
				token = strings.TrimSpace(
					cookie.Value,
				)
			}
		}

		// ========================================================
		// 3. Token Required
		// ========================================================

		if token == "" {

			http.Error(
				w,
				"Unauthorized",
				http.StatusUnauthorized,
			)

			return
		}

		// ========================================================
		// 4. Validate INTERNAL NID Session
		// ========================================================

		userID, valid :=
			helpers.ValidateInternalSessionToken(
				token,
			)

		if !valid {

			http.Error(
				w,
				"Invalid or expired NID session",
				http.StatusUnauthorized,
			)

			return
		}

		// ========================================================
		// 5. Store User ID in Context
		// ========================================================

		ctx := context.WithValue(
			r.Context(),
			UserIDKey,
			userID,
		)

		// ========================================================
		// 6. Continue
		// ========================================================

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}
