package handlers

import (
	"net/http"

	"mesascore/internal/middleware"
	"mesascore/internal/stats"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsHandler struct {
	pool *pgxpool.Pool
}

func NewStatsHandler(pool *pgxpool.Pool) *StatsHandler {
	return &StatsHandler{pool: pool}
}

func (h *StatsHandler) PartyLeaderboard(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())

	entries, err := stats.Leaderboard(r.Context(), h.pool, partyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	result := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		result = append(result, map[string]interface{}{
			"user": map[string]interface{}{
				"id":           e.UserID,
				"display_name": e.DisplayName,
				"avatar_url":   e.AvatarURL,
			},
			"wins":     e.Wins,
			"sessions": e.Sessions,
			"win_rate": e.WinRate,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *StatsHandler) PartyGame(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	gameID := chi.URLParam(r, "gameId")

	result, err := stats.GameStats(r.Context(), h.pool, partyID, gameID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *StatsHandler) PartyActivity(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())

	months, err := stats.ActivityPerMonth(r.Context(), h.pool, partyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, months)
}

func (h *StatsHandler) UserGlobal(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	result, err := stats.UserStats(r.Context(), h.pool, userID, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *StatsHandler) UserInParty(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())
	userID := chi.URLParam(r, "userId")

	result, err := stats.UserStats(r.Context(), h.pool, userID, &partyID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}
