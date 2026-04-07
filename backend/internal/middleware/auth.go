package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"mesascore/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.ParseJWT(tokenStr, jwtSecret)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Token refresh: if expiry within 24 hours, issue new token
			if claims.ExpiresAt != nil {
				timeUntilExpiry := time.Until(claims.ExpiresAt.Time)
				if timeUntilExpiry > 0 && timeUntilExpiry < 24*time.Hour {
					newToken, err := auth.IssueJWT(claims.Subject, jwtSecret, 7*24*time.Hour)
					if err == nil {
						w.Header().Set("X-New-Token", newToken)
					}
				}
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
