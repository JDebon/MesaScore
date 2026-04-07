package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"mesascore/internal/middleware"
	"mesascore/internal/validator"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionHandler struct {
	pool *pgxpool.Pool
}

func NewSessionHandler(pool *pgxpool.Pool) *SessionHandler {
	return &SessionHandler{pool: pool}
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())

	// Query params
	gameID := r.URL.Query().Get("game_id")
	userID := r.URL.Query().Get("user_id")
	sessionType := r.URL.Query().Get("type")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Build query
	query := `SELECT s.id, g.id, g.name, g.cover_image_url, s.session_type, s.played_at, s.duration_minutes,
	                  (SELECT COUNT(*) FROM session_participants sp WHERE sp.session_id = s.id) AS participant_count
	           FROM sessions s JOIN games g ON g.id = s.game_id
	           WHERE s.party_id = $1`
	countQuery := `SELECT COUNT(*) FROM sessions s WHERE s.party_id = $1`
	args := []interface{}{partyID}
	countArgs := []interface{}{partyID}
	argIdx := 2

	if gameID != "" {
		query += " AND s.game_id = $" + strconv.Itoa(argIdx)
		countQuery += " AND s.game_id = $" + strconv.Itoa(argIdx)
		args = append(args, gameID)
		countArgs = append(countArgs, gameID)
		argIdx++
	}
	if userID != "" {
		query += " AND EXISTS(SELECT 1 FROM session_participants sp WHERE sp.session_id = s.id AND sp.user_id = $" + strconv.Itoa(argIdx) + ")"
		countQuery += " AND EXISTS(SELECT 1 FROM session_participants sp WHERE sp.session_id = s.id AND sp.user_id = $" + strconv.Itoa(argIdx) + ")"
		args = append(args, userID)
		countArgs = append(countArgs, userID)
		argIdx++
	}
	if sessionType != "" {
		query += " AND s.session_type = $" + strconv.Itoa(argIdx) + "::session_type"
		countQuery += " AND s.session_type = $" + strconv.Itoa(argIdx) + "::session_type"
		args = append(args, sessionType)
		countArgs = append(countArgs, sessionType)
		argIdx++
	}
	if from != "" {
		query += " AND s.played_at >= $" + strconv.Itoa(argIdx)
		countQuery += " AND s.played_at >= $" + strconv.Itoa(argIdx)
		args = append(args, from)
		countArgs = append(countArgs, from)
		argIdx++
	}
	if to != "" {
		query += " AND s.played_at <= $" + strconv.Itoa(argIdx)
		countQuery += " AND s.played_at <= $" + strconv.Itoa(argIdx)
		args = append(args, to)
		countArgs = append(countArgs, to)
		argIdx++
	}

	// Total count
	var total int
	h.pool.QueryRow(r.Context(), countQuery, countArgs...).Scan(&total)

	query += " ORDER BY s.played_at DESC LIMIT $" + strconv.Itoa(argIdx) + " OFFSET $" + strconv.Itoa(argIdx+1)
	args = append(args, perPage, offset)

	rows, err := h.pool.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer rows.Close()

	sessions := []map[string]interface{}{}
	for rows.Next() {
		var sID, gID, gName, sType string
		var gCover *string
		var playedAt time.Time
		var duration *int
		var participantCount int
		rows.Scan(&sID, &gID, &gName, &gCover, &sType, &playedAt, &duration, &participantCount)

		// Get winners for this session
		winners := h.getWinners(r, sID, sType)

		sessions = append(sessions, map[string]interface{}{
			"id":   sID,
			"game": map[string]interface{}{"id": gID, "name": gName, "cover_image_url": gCover},
			"session_type":      sType,
			"played_at":         playedAt,
			"duration_minutes":  duration,
			"winners":           winners,
			"participant_count": participantCount,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":     sessions,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *SessionHandler) getWinners(r *http.Request, sessionID, sessionType string) []map[string]interface{} {
	var winnerQuery string
	switch sessionType {
	case "competitive", "score":
		winnerQuery = `SELECT u.id, u.display_name, u.avatar_url FROM session_participants sp
		               JOIN users u ON u.id = sp.user_id WHERE sp.session_id = $1 AND sp.rank = 1`
	case "team":
		winnerQuery = `SELECT u.id, u.display_name, u.avatar_url FROM session_participants sp
		               JOIN users u ON u.id = sp.user_id WHERE sp.session_id = $1 AND sp.rank = 1`
	case "cooperative":
		winnerQuery = `SELECT u.id, u.display_name, u.avatar_url FROM session_participants sp
		               JOIN users u ON u.id = sp.user_id WHERE sp.session_id = $1 AND sp.result = 'win'`
	default:
		return []map[string]interface{}{}
	}

	rows, err := h.pool.Query(r.Context(), winnerQuery, sessionID)
	if err != nil {
		return []map[string]interface{}{}
	}
	defer rows.Close()

	winners := []map[string]interface{}{}
	for rows.Next() {
		var id, displayName string
		var avatarURL *string
		rows.Scan(&id, &displayName, &avatarURL)
		winners = append(winners, map[string]interface{}{
			"id":           id,
			"display_name": displayName,
			"avatar_url":   avatarURL,
		})
	}
	return winners
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	partyID := middleware.PartyIDFromContext(r.Context())

	var sID, gID, gName, sType, createdByID, createdByName string
	var gCover *string
	var playedAt, createdAt time.Time
	var duration *int
	var notes *string
	var broughtByID *string
	var broughtByName *string

	err := h.pool.QueryRow(r.Context(),
		`SELECT s.id, g.id, g.name, g.cover_image_url, s.session_type, s.played_at,
		        s.duration_minutes, s.notes,
		        s.brought_by_user_id, bu.display_name,
		        s.created_by, cu.display_name, s.created_at
		 FROM sessions s
		 JOIN games g ON g.id = s.game_id
		 LEFT JOIN users bu ON bu.id = s.brought_by_user_id
		 JOIN users cu ON cu.id = s.created_by
		 WHERE s.id = $1 AND s.party_id = $2`, sessionID, partyID).
		Scan(&sID, &gID, &gName, &gCover, &sType, &playedAt,
			&duration, &notes, &broughtByID, &broughtByName, &createdByID, &createdByName, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	// Participants
	pRows, _ := h.pool.Query(r.Context(),
		`SELECT u.id, u.display_name, u.avatar_url, sp.team_name, sp.rank, sp.score, sp.result
		 FROM session_participants sp JOIN users u ON u.id = sp.user_id
		 WHERE sp.session_id = $1 ORDER BY sp.rank NULLS LAST, sp.score DESC NULLS LAST`, sessionID)
	defer pRows.Close()

	participants := []map[string]interface{}{}
	for pRows.Next() {
		var uID, uDisplayName string
		var uAvatarURL, teamName *string
		var rank *int
		var score *float64
		var result *string
		pRows.Scan(&uID, &uDisplayName, &uAvatarURL, &teamName, &rank, &score, &result)
		participants = append(participants, map[string]interface{}{
			"user":      map[string]interface{}{"id": uID, "display_name": uDisplayName, "avatar_url": uAvatarURL},
			"team_name": teamName,
			"rank":      rank,
			"score":     score,
			"result":    result,
		})
	}

	response := map[string]interface{}{
		"id":               sID,
		"game":             map[string]interface{}{"id": gID, "name": gName, "cover_image_url": gCover},
		"session_type":     sType,
		"played_at":        playedAt,
		"duration_minutes": duration,
		"notes":            notes,
		"created_by":       map[string]interface{}{"id": createdByID, "display_name": createdByName},
		"created_at":       createdAt,
		"participants":     participants,
	}
	if broughtByID != nil {
		response["brought_by"] = map[string]interface{}{"id": *broughtByID, "display_name": *broughtByName}
	} else {
		response["brought_by"] = nil
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	userID := middleware.UserIDFromContext(r.Context())

	var input validator.SessionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if validationErrors := validator.ValidateSession(r.Context(), h.pool, partyID, input); len(validationErrors) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "validation failed", "fields": validationErrors})
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer tx.Rollback(r.Context())

	var sessionID string
	err = tx.QueryRow(r.Context(),
		`INSERT INTO sessions (party_id, game_id, session_type, played_at, duration_minutes, brought_by_user_id, notes, created_by)
		 VALUES ($1, $2, $3::session_type, $4, $5, $6, $7, $8) RETURNING id`,
		partyID, input.GameID, input.SessionType, input.PlayedAt, input.DurationMinutes,
		input.BroughtByUserID, input.Notes, userID).Scan(&sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create session"})
		return
	}

	for _, p := range input.Participants {
		_, err = tx.Exec(r.Context(),
			`INSERT INTO session_participants (session_id, user_id, team_name, rank, score, result)
			 VALUES ($1, $2, $3, $4, $5, $6::participant_result)`,
			sessionID, p.UserID, p.TeamName, p.Rank, p.Score, p.Result)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add participant"})
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to commit"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": sessionID})
}

func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	sessionID := chi.URLParam(r, "sessionId")

	// Verify session exists in this party
	var exists bool
	h.pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM sessions WHERE id = $1 AND party_id = $2)",
		sessionID, partyID).Scan(&exists)
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}

	var input validator.SessionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if validationErrors := validator.ValidateSession(r.Context(), h.pool, partyID, input); len(validationErrors) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"error": "validation failed", "fields": validationErrors})
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer tx.Rollback(r.Context())

	_, err = tx.Exec(r.Context(),
		`UPDATE sessions SET game_id = $1, session_type = $2::session_type, played_at = $3,
		 duration_minutes = $4, brought_by_user_id = $5, notes = $6 WHERE id = $7`,
		input.GameID, input.SessionType, input.PlayedAt, input.DurationMinutes,
		input.BroughtByUserID, input.Notes, sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update session"})
		return
	}

	// Replace participants
	tx.Exec(r.Context(), "DELETE FROM session_participants WHERE session_id = $1", sessionID)
	for _, p := range input.Participants {
		_, err = tx.Exec(r.Context(),
			`INSERT INTO session_participants (session_id, user_id, team_name, rank, score, result)
			 VALUES ($1, $2, $3, $4, $5, $6::participant_result)`,
			sessionID, p.UserID, p.TeamName, p.Rank, p.Score, p.Result)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update participants"})
			return
		}
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to commit"})
		return
	}

	// Return updated session detail
	h.Get(w, r)
}

func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	sessionID := chi.URLParam(r, "sessionId")

	tag, err := h.pool.Exec(r.Context(),
		"DELETE FROM sessions WHERE id = $1 AND party_id = $2", sessionID, partyID)
	if err != nil || tag.RowsAffected() == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
