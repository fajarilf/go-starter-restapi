package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/fajarilf/go-starter-api/internal/service"
)

type contextKey string

const userClaimsKey contextKey = "userClaims"

// UserClaimsFromContext extracts JWT claims from a request context.
func UserClaimsFromContext(ctx context.Context) *domain.JWTClaims {
	claims, _ := ctx.Value(userClaimsKey).(*domain.JWTClaims)
	return claims
}

// RequireAuth is a chi-compatible middleware that validates the JWT Bearer
// token and injects the parsed claims into the request context.
func RequireAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"missing or invalid token"}`))
				return
			}

			claims, err := auth.ValidateToken(token)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid token"}`))
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
