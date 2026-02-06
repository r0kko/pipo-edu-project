package auth

import (
	"net/http"
	"strings"

	"pipo-edu-project/internal/identity"
)

func AuthMiddleware(tokens *TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing Authorization", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid Authorization", http.StatusUnauthorized)
				return
			}
			claims, err := tokens.ParseAccess(parts[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := identity.WithUser(r.Context(), claims.UserID, string(claims.Role))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRoles(allowed ...Role) func(http.Handler) http.Handler {
	allowedSet := map[string]struct{}{}
	for _, role := range allowed {
		allowedSet[string(role)] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := identity.RoleFrom(r.Context())
			if !ok {
				http.Error(w, "missing role", http.StatusForbidden)
				return
			}
			if _, exists := allowedSet[role]; !exists {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
