package api

import (
	"context"
	"net/http"
	"strings"

	"syncspace/backend/internal/auth"
)

type contextKey string

const userContextKey contextKey = "user"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, 401, "unauthorized", "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeError(w, 401, "unauthorized", "invalid authorization header format")
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			writeError(w, 401, "unauthorized", "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(userContextKey).(*auth.Claims)
			if !ok {
				writeError(w, 401, "unauthorized", "authentication required")
				return
			}

			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeError(w, 403, "forbidden", "insufficient permissions")
		})
	}
}

func GetUserFromContext(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(userContextKey).(*auth.Claims)
	return claims
}
