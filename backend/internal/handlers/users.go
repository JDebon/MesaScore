package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"mesascore/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	pool *pgxpool.Pool
}

func NewUserHandler(pool *pgxpool.Pool) *UserHandler {
	return &UserHandler{pool: pool}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var id, username, displayName string
	var avatarURL *string
	var createdAt time.Time

	err := h.pool.QueryRow(r.Context(),
		"SELECT id, username, display_name, avatar_url, created_at FROM users WHERE id = $1",
		userID).Scan(&id, &username, &displayName, &avatarURL, &createdAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           id,
		"username":     username,
		"display_name": displayName,
		"avatar_url":   avatarURL,
		"created_at":   createdAt,
	})
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		DisplayName *string `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.DisplayName != nil {
		h.pool.Exec(r.Context(), "UPDATE users SET display_name = $1 WHERE id = $2", *req.DisplayName, userID)
	}
	if req.AvatarURL != nil {
		h.pool.Exec(r.Context(), "UPDATE users SET avatar_url = $1 WHERE id = $2", *req.AvatarURL, userID)
	}

	h.getUser(w, r, userID)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.getUser(w, r, id)
}

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request, userID string) {
	var id, username, displayName string
	var avatarURL *string
	var createdAt time.Time

	err := h.pool.QueryRow(r.Context(),
		"SELECT id, username, display_name, avatar_url, created_at FROM users WHERE id = $1",
		userID).Scan(&id, &username, &displayName, &avatarURL, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		log.Printf("getUser error for userID=%q: %v", userID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           id,
		"username":     username,
		"display_name": displayName,
		"avatar_url":   avatarURL,
		"created_at":   createdAt,
	})
}

func (h *UserHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	userID := middleware.UserIDFromContext(r.Context())
	pattern := "%" + query + "%"

	rows, err := h.pool.Query(r.Context(),
		`SELECT id, username, display_name, avatar_url FROM users
		 WHERE id != $1 AND (username ILIKE $2 OR display_name ILIKE $2)
		 LIMIT 10`,
		userID, pattern)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	for rows.Next() {
		var id, username, displayName string
		var avatarURL *string
		rows.Scan(&id, &username, &displayName, &avatarURL)
		results = append(results, map[string]interface{}{
			"id":           id,
			"username":     username,
			"display_name": displayName,
			"avatar_url":   avatarURL,
		})
	}

	writeJSON(w, http.StatusOK, results)
}
