package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/library_system/internal/config"
	"github.com/library_system/internal/storage"
	"github.com/library_system/internal/utils/response"
	"github.com/library_system/utils"
)

// AdminOnly must be chained after AuthMiddleware (which sets "userRole" in context).
// It returns 403 if the authenticated user does not have the "admin" role.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value("userRole").(string)
		if role != "admin" {
			response.ApiErrorResponse(w, http.StatusForbidden, errors.New("admin access required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(cfg *config.Config, db storage.Storage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var token string

			// 1️⃣ Try Authorization header first
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || parts[0] != "Bearer" {
					response.ApiErrorResponse(w, http.StatusUnauthorized, nil)
					return
				}
				token = parts[1]
			} else {
				// 2️⃣ Try cookie as fallback
				cookie, err := r.Cookie("auth_token")
				if err != nil {
					response.ApiErrorResponse(w, http.StatusUnauthorized, nil)
					return
				}
				token = cookie.Value
			}

			// 3️⃣ Check token validity
			if token == "" {
				response.ApiErrorResponse(w, http.StatusUnauthorized, nil)
				return
			}

			// 4️⃣ Verify JWT token with secret and extract userId + role
			userId, role, err := utils.VerifyJWT(token, cfg.JWT.Secret)
			if err != nil {
				response.ApiErrorResponse(w, http.StatusUnauthorized, nil)
				return
			}

			// 5️⃣ Verify user still exists in database (efficient check)
			exists, err := db.UserExists(userId)
			if err != nil || !exists {
				response.ApiErrorResponse(w, http.StatusUnauthorized, errors.New("user not found"))
				return
			}

			// 6️⃣ Pass userId and role to handlers via context
			ctx := context.WithValue(r.Context(), "userId", userId)
			ctx = context.WithValue(ctx, "userRole", role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
