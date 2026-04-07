package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"mesascore/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CollectionHandler struct {
	pool *pgxpool.Pool
}

func NewCollectionHandler(pool *pgxpool.Pool) *CollectionHandler {
	return &CollectionHandler{pool: pool}
}

func (h *CollectionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	rows, err := h.pool.Query(r.Context(),
		`SELECT g.id, g.name, g.cover_image_url, ug.added_at
		 FROM user_games ug JOIN games g ON g.id = ug.game_id
		 WHERE ug.user_id = $1
		 ORDER BY g.name`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	games := []map[string]interface{}{}
	for rows.Next() {
		var gameID, name string
		var coverImageURL *string
		var addedAt time.Time
		rows.Scan(&gameID, &name, &coverImageURL, &addedAt)
		games = append(games, map[string]interface{}{
			"game_id":         gameID,
			"name":            name,
			"cover_image_url": coverImageURL,
			"added_at":        addedAt,
		})
	}

	writeJSON(w, http.StatusOK, games)
}

func (h *CollectionHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		GameID string `json:"game_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Check game exists
	var exists bool
	h.pool.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM games WHERE id = $1)", req.GameID).Scan(&exists)
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "game not found"})
		return
	}

	// Check not already in collection
	h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM user_games WHERE user_id = $1 AND game_id = $2)",
		userID, req.GameID).Scan(&exists)
	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "game already in collection"})
		return
	}

	_, err := h.pool.Exec(r.Context(),
		"INSERT INTO user_games (user_id, game_id) VALUES ($1, $2)", userID, req.GameID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add game"})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *CollectionHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	gameID := chi.URLParam(r, "gameId")

	tag, err := h.pool.Exec(r.Context(),
		"DELETE FROM user_games WHERE user_id = $1 AND game_id = $2", userID, gameID)
	if err != nil || tag.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "game not in collection"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
