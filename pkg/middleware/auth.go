package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/suleymankursatdemir/ecommerce-platform/internal/auth/domain"
	"github.com/suleymankursatdemir/ecommerce-platform/internal/auth/usecase"
)

type authContextKey string

const (
	UserIDKey authContextKey = "user_id"
	RoleKey   authContextKey = "user_role"
)

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

func GetUserRole(ctx context.Context) domain.Role {
	if role, ok := ctx.Value(RoleKey).(domain.Role); ok {
		return role
	}
	return ""
}

func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeAuthError(w, "invalid authorization header format")
				return
			}

			tokenString := parts[1]
			claims, err := usecase.ParseAccessToken(tokenString, jwtSecret)
			if err != nil {
				writeAuthError(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	roleSet := make(map[domain.Role]bool, len(roles))
	for _, role := range roles {
		roleSet[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if !roleSet[userRole] {
				writeAuthError(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
