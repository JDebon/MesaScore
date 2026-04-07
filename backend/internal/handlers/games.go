package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"mesascore/internal/bgg"
	"mesascore/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameHandler struct {
	pool *pgxpool.Pool
}

func NewGameHandler(pool *pgxpool.Pool) *GameHandler {
	return &GameHandler{pool: pool}
}

func (h *GameHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	q := r.URL.Query().Get("q")
	sort := r.URL.Query().Get("sort")

	query := `SELECT g.id, g.bgg_id, g.name, g.cover_image_url, g.min_players, g.max_players, g.bgg_rating,
	                  (SELECT COUNT(*) FROM sessions s WHERE s.game_id = g.id) AS session_count,
	                  EXISTS(SELECT 1 FROM user_games ug WHERE ug.game_id = g.id AND ug.user_id = $1) AS in_my_collection
	           FROM games g`

	args := []interface{}{userID}
	if q != "" {
		query += " WHERE g.name ILIKE $2"
		args = append(args, "%"+q+"%")
	}

	switch sort {
	case "rating":
		query += " ORDER BY g.bgg_rating DESC NULLS LAST"
	case "sessions":
		query += " ORDER BY session_count DESC"
	default:
		query += " ORDER BY g.name"
	}

	rows, err := h.pool.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	games := []map[string]interface{}{}
	for rows.Next() {
		var id, name string
		var bggID *int
		var coverImageURL *string
		var minPlayers, maxPlayers *int
		var bggRating *float64
		var sessionCount int
		var inMyCollection bool
		rows.Scan(&id, &bggID, &name, &coverImageURL, &minPlayers, &maxPlayers, &bggRating, &sessionCount, &inMyCollection)
		games = append(games, map[string]interface{}{
			"id":               id,
			"bgg_id":           bggID,
			"name":             name,
			"cover_image_url":  coverImageURL,
			"min_players":      minPlayers,
			"max_players":      maxPlayers,
			"bgg_rating":       bggRating,
			"session_count":    sessionCount,
			"in_my_collection": inMyCollection,
		})
	}

	writeJSON(w, http.StatusOK, games)
}

func (h *GameHandler) Get(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")
	userID := middleware.UserIDFromContext(r.Context())

	var id, name string
	var bggID *int
	var description, coverImageURL *string
	var minPlayers, maxPlayers *int
	var bggRating *float64
	var bggFetchedAt *time.Time
	var addedByID, addedByDisplayName string
	var createdAt time.Time

	err := h.pool.QueryRow(r.Context(),
		`SELECT g.id, g.bgg_id, g.name, g.description, g.cover_image_url,
		        g.min_players, g.max_players, g.bgg_rating, g.bgg_fetched_at,
		        g.added_by_user_id, u.display_name, g.created_at
		 FROM games g JOIN users u ON u.id = g.added_by_user_id
		 WHERE g.id = $1`, gameID).
		Scan(&id, &bggID, &name, &description, &coverImageURL,
			&minPlayers, &maxPlayers, &bggRating, &bggFetchedAt,
			&addedByID, &addedByDisplayName, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "game not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	// Session count
	var sessionCount int
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM sessions WHERE game_id = $1", gameID).Scan(&sessionCount)

	// In collection
	var inMyCollection bool
	h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM user_games WHERE game_id = $1 AND user_id = $2)",
		gameID, userID).Scan(&inMyCollection)

	// Owners
	ownerRows, _ := h.pool.Query(r.Context(),
		`SELECT u.id, u.display_name, u.avatar_url
		 FROM user_games ug JOIN users u ON u.id = ug.user_id
		 WHERE ug.game_id = $1`, gameID)
	defer ownerRows.Close()

	owners := []map[string]interface{}{}
	for ownerRows.Next() {
		var oID, oDisplayName string
		var oAvatarURL *string
		ownerRows.Scan(&oID, &oDisplayName, &oAvatarURL)
		owners = append(owners, map[string]interface{}{
			"id":           oID,
			"display_name": oDisplayName,
			"avatar_url":   oAvatarURL,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":               id,
		"bgg_id":           bggID,
		"name":             name,
		"description":      description,
		"cover_image_url":  coverImageURL,
		"min_players":      minPlayers,
		"max_players":      maxPlayers,
		"bgg_rating":       bggRating,
		"bgg_fetched_at":   bggFetchedAt,
		"added_by":         map[string]interface{}{"id": addedByID, "display_name": addedByDisplayName},
		"created_at":       createdAt,
		"session_count":    sessionCount,
		"in_my_collection": inMyCollection,
		"owners":           owners,
	})
}

func (h *GameHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		BggID *int    `json:"bgg_id"`
		Name  *string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.BggID == nil && (req.Name == nil || *req.Name == "") {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "validation failed",
			"fields": map[string]string{"bgg_id": "bgg_id or name required"},
		})
		return
	}

	var (
		name          string
		description   *string
		coverImageURL *string
		minPlayers    *int
		maxPlayers    *int
		bggRating     *float64
		bggFetchedAt  *time.Time
	)

	if req.BggID != nil {
		// Check duplicate
		var exists bool
		h.pool.QueryRow(r.Context(),
			"SELECT EXISTS(SELECT 1 FROM games WHERE bgg_id = $1)", *req.BggID).Scan(&exists)
		if exists {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "game with this BGG ID already exists"})
			return
		}

		detail, err := bgg.GetThing(*req.BggID)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "BGG unreachable"})
			return
		}
		now := time.Now()
		name = detail.Name
		description = detail.Description
		coverImageURL = detail.CoverImageURL
		minPlayers = detail.MinPlayers
		maxPlayers = detail.MaxPlayers
		bggRating = detail.BggRating
		bggFetchedAt = &now
	} else {
		name = *req.Name
	}

	var gameID string
	err := h.pool.QueryRow(r.Context(),
		`INSERT INTO games (bgg_id, name, description, cover_image_url, min_players, max_players, bgg_rating, bgg_fetched_at, added_by_user_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		req.BggID, name, description, coverImageURL, minPlayers, maxPlayers, bggRating, bggFetchedAt, userID).
		Scan(&gameID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create game"})
		return
	}

	// Auto-add to user's collection
	h.pool.Exec(r.Context(),
		"INSERT INTO user_games (user_id, game_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		userID, gameID)

	writeJSON(w, http.StatusCreated, map[string]string{"id": gameID})
}

func (h *GameHandler) Update(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")
	userID := middleware.UserIDFromContext(r.Context())

	// Check authorization: must own the game or have added it
	var exists bool
	h.pool.QueryRow(r.Context(),
		`SELECT EXISTS(
			SELECT 1 FROM games WHERE id = $1 AND added_by_user_id = $2
			UNION
			SELECT 1 FROM user_games WHERE game_id = $1 AND user_id = $2
		)`, gameID, userID).Scan(&exists)
	if !exists {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	var req struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		CoverImageURL *string `json:"cover_image_url"`
		MinPlayers    *int    `json:"min_players"`
		MaxPlayers    *int    `json:"max_players"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name != nil {
		h.pool.Exec(r.Context(), "UPDATE games SET name = $1 WHERE id = $2", *req.Name, gameID)
	}
	if req.Description != nil {
		h.pool.Exec(r.Context(), "UPDATE games SET description = $1 WHERE id = $2", *req.Description, gameID)
	}
	if req.CoverImageURL != nil {
		h.pool.Exec(r.Context(), "UPDATE games SET cover_image_url = $1 WHERE id = $2", *req.CoverImageURL, gameID)
	}
	if req.MinPlayers != nil {
		h.pool.Exec(r.Context(), "UPDATE games SET min_players = $1 WHERE id = $2", *req.MinPlayers, gameID)
	}
	if req.MaxPlayers != nil {
		h.pool.Exec(r.Context(), "UPDATE games SET max_players = $1 WHERE id = $2", *req.MaxPlayers, gameID)
	}

	// Return updated game
	h.Get(w, r)
}

func (h *GameHandler) BGGRefresh(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")
	userID := middleware.UserIDFromContext(r.Context())

	// Check user owns it
	var exists bool
	h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM user_games WHERE game_id = $1 AND user_id = $2)",
		gameID, userID).Scan(&exists)
	if !exists {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	var bggID *int
	h.pool.QueryRow(r.Context(), "SELECT bgg_id FROM games WHERE id = $1", gameID).Scan(&bggID)
	if bggID == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "game has no BGG ID"})
		return
	}

	detail, err := bgg.GetThing(*bggID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "BGG unreachable"})
		return
	}

	now := time.Now()
	// Name is NOT overwritten on refresh per spec
	h.pool.Exec(r.Context(),
		`UPDATE games SET description = $1, cover_image_url = $2, bgg_rating = $3,
		 min_players = $4, max_players = $5, bgg_fetched_at = $6 WHERE id = $7`,
		detail.Description, detail.CoverImageURL, detail.BggRating,
		detail.MinPlayers, detail.MaxPlayers, now, gameID)

	h.Get(w, r)
}

func (h *GameHandler) AvailableForParty(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())

	rows, err := h.pool.Query(r.Context(),
		`SELECT DISTINCT g.id, g.name, g.cover_image_url
		 FROM games g
		 JOIN user_games ug ON ug.game_id = g.id
		 JOIN party_members pm ON pm.user_id = ug.user_id AND pm.party_id = $1
		 ORDER BY g.name`, partyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	type gameWithOwners struct {
		ID            string                   `json:"id"`
		Name          string                   `json:"name"`
		CoverImageURL *string                  `json:"cover_image_url"`
		Owners        []map[string]interface{} `json:"owners"`
	}

	gamesMap := map[string]*gameWithOwners{}
	var orderedIDs []string
	for rows.Next() {
		var id, name string
		var coverImageURL *string
		rows.Scan(&id, &name, &coverImageURL)
		if _, ok := gamesMap[id]; !ok {
			gamesMap[id] = &gameWithOwners{ID: id, Name: name, CoverImageURL: coverImageURL, Owners: []map[string]interface{}{}}
			orderedIDs = append(orderedIDs, id)
		}
	}

	// Fetch owners for each game in party
	for _, gID := range orderedIDs {
		ownerRows, _ := h.pool.Query(r.Context(),
			`SELECT u.id, u.display_name FROM users u
			 JOIN user_games ug ON ug.user_id = u.id AND ug.game_id = $1
			 JOIN party_members pm ON pm.user_id = u.id AND pm.party_id = $2`,
			gID, partyID)
		for ownerRows.Next() {
			var oID, oName string
			ownerRows.Scan(&oID, &oName)
			gamesMap[gID].Owners = append(gamesMap[gID].Owners, map[string]interface{}{
				"id": oID, "display_name": oName,
			})
		}
		ownerRows.Close()
	}

	result := make([]*gameWithOwners, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		result = append(result, gamesMap[id])
	}

	writeJSON(w, http.StatusOK, result)
}

type BGGHandler struct{}

func NewBGGHandler() *BGGHandler {
	return &BGGHandler{}
}

func (h *BGGHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	results, err := bgg.Search(q)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "BGG unreachable"})
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *BGGHandler) Thing(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id must be an integer"})
		return
	}

	detail, err := bgg.GetThing(id)
	if err != nil {
		if err.Error() == "game not found on BGG" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "game not found on BGG"})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "BGG unreachable"})
		return
	}

	writeJSON(w, http.StatusOK, detail)
}
