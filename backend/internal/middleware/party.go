package middleware

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const PartyIDKey contextKey = "partyID"

func PartyIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(PartyIDKey).(string)
	return id
}

func PartyMemberMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			partyID := chi.URLParam(r, "partyId")
			userID := UserIDFromContext(r.Context())

			// Check party exists
			var exists bool
			err := pool.QueryRow(r.Context(),
				"SELECT EXISTS(SELECT 1 FROM parties WHERE id = $1)", partyID).Scan(&exists)
			if err != nil || !exists {
				log.Printf("PartyMemberMiddleware: party check failed for partyID=%q, exists=%v, err=%v", partyID, exists, err)
				http.Error(w, `{"error":"party not found"}`, http.StatusNotFound)
				return
			}

			// Check membership
			err = pool.QueryRow(r.Context(),
				"SELECT EXISTS(SELECT 1 FROM party_members WHERE party_id = $1 AND user_id = $2)",
				partyID, userID).Scan(&exists)
			if err != nil || !exists {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), PartyIDKey, partyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func PartyAdminMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			partyID := chi.URLParam(r, "partyId")
			userID := UserIDFromContext(r.Context())

			var adminID string
			err := pool.QueryRow(r.Context(),
				"SELECT admin_user_id FROM parties WHERE id = $1", partyID).Scan(&adminID)
			if err != nil {
				http.Error(w, `{"error":"party not found"}`, http.StatusNotFound)
				return
			}

			if adminID != userID {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), PartyIDKey, partyID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
