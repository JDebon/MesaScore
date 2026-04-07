package handlers

import (
	"net/http"
	"time"

	"mesascore/internal/middleware"
	"mesascore/internal/stats"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardHandler struct {
	pool *pgxpool.Pool
}

func NewDashboardHandler(pool *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{pool: pool}
}

func (h *DashboardHandler) UserDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	// Parties
	partyRows, err := h.pool.Query(r.Context(),
		`SELECT p.id, p.name,
		        (SELECT COUNT(*) FROM party_members pm2 WHERE pm2.party_id = p.id) AS member_count,
		        (SELECT MAX(s.played_at) FROM sessions s WHERE s.party_id = p.id) AS last_session_at
		 FROM parties p
		 JOIN party_members pm ON pm.party_id = p.id
		 WHERE pm.user_id = $1
		 ORDER BY p.name`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer partyRows.Close()

	parties := []map[string]interface{}{}
	for partyRows.Next() {
		var pID, pName string
		var memberCount int
		var lastSessionAt *time.Time
		partyRows.Scan(&pID, &pName, &memberCount, &lastSessionAt)
		parties = append(parties, map[string]interface{}{
			"id":              pID,
			"name":            pName,
			"member_count":    memberCount,
			"last_session_at": lastSessionAt,
		})
	}

	// Pending invites
	inviteRows, _ := h.pool.Query(r.Context(),
		`SELECT pi.id, p.id, p.name, u.id, u.display_name, pi.created_at
		 FROM party_invites pi
		 JOIN parties p ON p.id = pi.party_id
		 JOIN users u ON u.id = pi.invited_by_user_id
		 WHERE pi.invited_user_id = $1 AND pi.status = 'pending'
		 ORDER BY pi.created_at DESC`, userID)
	defer inviteRows.Close()

	invites := []map[string]interface{}{}
	for inviteRows.Next() {
		var invID, pID, pName, uID, uName string
		var createdAt time.Time
		inviteRows.Scan(&invID, &pID, &pName, &uID, &uName, &createdAt)
		invites = append(invites, map[string]interface{}{
			"id":         invID,
			"party":      map[string]interface{}{"id": pID, "name": pName},
			"invited_by": map[string]interface{}{"id": uID, "display_name": uName},
			"created_at": createdAt,
		})
	}

	// Global stats
	var totalSessions, totalWins int
	h.pool.QueryRow(r.Context(),
		`SELECT COUNT(DISTINCT sp.session_id),
		        COUNT(DISTINCT sp.session_id) FILTER (WHERE sp.rank = 1 OR sp.result = 'win')
		 FROM session_participants sp
		 WHERE sp.user_id = $1`, userID).Scan(&totalSessions, &totalWins)

	currentStreak, _ := stats.CalculateStreaks(r.Context(), h.pool, userID, nil)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"parties":         parties,
		"pending_invites": invites,
		"global_stats": map[string]interface{}{
			"total_sessions": totalSessions,
			"total_wins":     totalWins,
			"current_streak": currentStreak,
		},
	})
}

func (h *DashboardHandler) PartyDashboard(w http.ResponseWriter, r *http.Request) {
	partyID := middleware.PartyIDFromContext(r.Context())

	// Totals
	var totalSessions, totalMembers int
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM sessions WHERE party_id = $1", partyID).Scan(&totalSessions)
	h.pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM party_members WHERE party_id = $1", partyID).Scan(&totalMembers)

	var totalUniqueGames int
	h.pool.QueryRow(r.Context(),
		"SELECT COUNT(DISTINCT game_id) FROM sessions WHERE party_id = $1", partyID).Scan(&totalUniqueGames)

	// Current leader
	var currentLeader *map[string]interface{}
	leaderRow := h.pool.QueryRow(r.Context(),
		`SELECT u.id, u.display_name, u.avatar_url, COUNT(*) AS wins
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 JOIN users u ON u.id = sp.user_id
		 WHERE s.party_id = $1 AND (sp.rank = 1 OR sp.result = 'win')
		 GROUP BY u.id, u.display_name, u.avatar_url
		 ORDER BY wins DESC LIMIT 1`, partyID)
	var lID, lName string
	var lAvatar *string
	var lWins int
	if err := leaderRow.Scan(&lID, &lName, &lAvatar, &lWins); err == nil {
		leader := map[string]interface{}{
			"user": map[string]interface{}{"id": lID, "display_name": lName, "avatar_url": lAvatar},
			"wins": lWins,
		}
		currentLeader = &leader
	}

	// Most played game
	var mostPlayedGame *map[string]interface{}
	mpRow := h.pool.QueryRow(r.Context(),
		`SELECT g.id, g.name, g.cover_image_url, COUNT(*) AS cnt
		 FROM sessions s JOIN games g ON g.id = s.game_id
		 WHERE s.party_id = $1
		 GROUP BY g.id, g.name, g.cover_image_url
		 ORDER BY cnt DESC LIMIT 1`, partyID)
	var mpID, mpName string
	var mpCover *string
	var mpCount int
	if err := mpRow.Scan(&mpID, &mpName, &mpCover, &mpCount); err == nil {
		mp := map[string]interface{}{"id": mpID, "name": mpName, "cover_image_url": mpCover, "session_count": mpCount}
		mostPlayedGame = &mp
	}

	// Sessions per month
	months, _ := stats.ActivityPerMonth(r.Context(), h.pool, partyID)

	// Recent sessions (last 5)
	recentRows, _ := h.pool.Query(r.Context(),
		`SELECT s.id, g.id, g.name, g.cover_image_url, s.played_at, s.session_type
		 FROM sessions s JOIN games g ON g.id = s.game_id
		 WHERE s.party_id = $1
		 ORDER BY s.played_at DESC LIMIT 5`, partyID)
	defer recentRows.Close()

	recentSessions := []map[string]interface{}{}
	for recentRows.Next() {
		var sID, gID, gName, sType string
		var playedAt time.Time
		var gCover *string
		recentRows.Scan(&sID, &gID, &gName, &gCover, &playedAt, &sType)

		// Winners
		var winnerQuery string
		switch sType {
		case "cooperative":
			winnerQuery = `SELECT u.id, u.display_name FROM session_participants sp JOIN users u ON u.id = sp.user_id WHERE sp.session_id = $1 AND sp.result = 'win'`
		default:
			winnerQuery = `SELECT u.id, u.display_name FROM session_participants sp JOIN users u ON u.id = sp.user_id WHERE sp.session_id = $1 AND sp.rank = 1`
		}
		wRows, _ := h.pool.Query(r.Context(), winnerQuery, sID)
		winners := []map[string]interface{}{}
		for wRows.Next() {
			var wID, wName string
			wRows.Scan(&wID, &wName)
			winners = append(winners, map[string]interface{}{"id": wID, "display_name": wName})
		}
		wRows.Close()

		recentSessions = append(recentSessions, map[string]interface{}{
			"id":           sID,
			"game":         map[string]interface{}{"id": gID, "name": gName, "cover_image_url": gCover},
			"played_at":    playedAt,
			"session_type": sType,
			"winners":      winners,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_sessions":     totalSessions,
		"total_unique_games": totalUniqueGames,
		"total_members":      totalMembers,
		"current_leader":     currentLeader,
		"most_played_game":   mostPlayedGame,
		"sessions_per_month": months,
		"recent_sessions":    recentSessions,
	})
}
