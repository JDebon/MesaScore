package stats

import (
	"context"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LeaderboardEntry struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Wins        int     `json:"wins"`
	Sessions    int     `json:"sessions"`
	WinRate     float64 `json:"win_rate"`
}

type MonthCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type GameRef struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
}

type PlayerRef struct {
	ID           string  `json:"id"`
	DisplayName  string  `json:"display_name"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	Count        int     `json:"count"` // wins_against or losses_against
}

type PerGameStat struct {
	Game     GameRef `json:"game"`
	Sessions int     `json:"sessions"`
	Wins     int     `json:"wins"`
	WinRate  float64 `json:"win_rate"`
}

type HeadToHeadEntry struct {
	Opponent        PlayerRef `json:"opponent"`
	SessionsTogether int      `json:"sessions_together"`
	ThisUserWins     int      `json:"this_user_wins"`
	OpponentWins     int      `json:"opponent_wins"`
}

// Leaderboard returns win/session/winrate per member of a party.
func Leaderboard(ctx context.Context, pool *pgxpool.Pool, partyID string) ([]LeaderboardEntry, error) {
	rows, err := pool.Query(ctx,
		`SELECT u.id, u.display_name, u.avatar_url,
		        COUNT(DISTINCT sp.session_id) AS sessions,
		        COUNT(DISTINCT sp.session_id) FILTER (
		            WHERE (sp.rank = 1) OR (sp.result = 'win')
		        ) AS wins
		 FROM party_members pm
		 JOIN users u ON u.id = pm.user_id
		 LEFT JOIN session_participants sp ON sp.user_id = pm.user_id
		 LEFT JOIN sessions s ON s.id = sp.session_id AND s.party_id = $1
		 WHERE pm.party_id = $1
		 GROUP BY u.id, u.display_name, u.avatar_url
		 ORDER BY wins DESC, sessions DESC`, partyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []LeaderboardEntry{}
	for rows.Next() {
		var e LeaderboardEntry
		rows.Scan(&e.UserID, &e.DisplayName, &e.AvatarURL, &e.Sessions, &e.Wins)
		if e.Sessions > 0 {
			e.WinRate = float64(e.Wins) / float64(e.Sessions)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// GameStats returns per-game leaderboard within a party.
func GameStats(ctx context.Context, pool *pgxpool.Pool, partyID, gameID string) (map[string]interface{}, error) {
	var gameName string
	var totalSessions int
	var lastPlayedAt *time.Time

	pool.QueryRow(ctx, "SELECT name FROM games WHERE id = $1", gameID).Scan(&gameName)
	pool.QueryRow(ctx,
		"SELECT COUNT(*), MAX(played_at) FROM sessions WHERE party_id = $1 AND game_id = $2",
		partyID, gameID).Scan(&totalSessions, &lastPlayedAt)

	// Sessions per month
	monthRows, _ := pool.Query(ctx,
		`SELECT TO_CHAR(played_at, 'YYYY-MM') AS month, COUNT(*) FROM sessions
		 WHERE party_id = $1 AND game_id = $2
		 GROUP BY month ORDER BY month DESC LIMIT 12`, partyID, gameID)
	defer monthRows.Close()

	months := []MonthCount{}
	for monthRows.Next() {
		var mc MonthCount
		monthRows.Scan(&mc.Month, &mc.Count)
		months = append(months, mc)
	}

	// Leaderboard for this game
	lbRows, _ := pool.Query(ctx,
		`SELECT u.id, u.display_name, u.avatar_url,
		        COUNT(DISTINCT sp.session_id) AS sessions,
		        COUNT(DISTINCT sp.session_id) FILTER (WHERE sp.rank = 1 OR sp.result = 'win') AS wins
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 JOIN users u ON u.id = sp.user_id
		 WHERE s.party_id = $1 AND s.game_id = $2
		 GROUP BY u.id, u.display_name, u.avatar_url
		 ORDER BY wins DESC`, partyID, gameID)
	defer lbRows.Close()

	leaderboard := []LeaderboardEntry{}
	for lbRows.Next() {
		var e LeaderboardEntry
		lbRows.Scan(&e.UserID, &e.DisplayName, &e.AvatarURL, &e.Sessions, &e.Wins)
		if e.Sessions > 0 {
			e.WinRate = float64(e.Wins) / float64(e.Sessions)
		}
		leaderboard = append(leaderboard, e)
	}

	return map[string]interface{}{
		"game":               map[string]interface{}{"id": gameID, "name": gameName},
		"total_sessions":     totalSessions,
		"last_played_at":     lastPlayedAt,
		"sessions_per_month": months,
		"leaderboard":        leaderboard,
	}, nil
}

// ActivityPerMonth returns session counts per month for a party.
func ActivityPerMonth(ctx context.Context, pool *pgxpool.Pool, partyID string) ([]MonthCount, error) {
	rows, err := pool.Query(ctx,
		`SELECT TO_CHAR(played_at, 'YYYY-MM') AS month, COUNT(*) FROM sessions
		 WHERE party_id = $1
		 GROUP BY month ORDER BY month DESC LIMIT 12`, partyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	months := []MonthCount{}
	for rows.Next() {
		var mc MonthCount
		rows.Scan(&mc.Month, &mc.Count)
		months = append(months, mc)
	}
	return months, nil
}

// UserStats returns full stats for a user, optionally scoped to a party.
func UserStats(ctx context.Context, pool *pgxpool.Pool, userID string, partyID *string) (map[string]interface{}, error) {
	// Base filter
	partyFilter := ""
	args := []interface{}{userID}
	if partyID != nil {
		partyFilter = " AND s.party_id = $2"
		args = append(args, *partyID)
	}

	// Total sessions and wins
	var totalSessions, totalWins int
	pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT sp.session_id),
		        COUNT(DISTINCT sp.session_id) FILTER (WHERE sp.rank = 1 OR sp.result = 'win')
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 WHERE sp.user_id = $1`+partyFilter, args...).Scan(&totalSessions, &totalWins)

	var winRate float64
	if totalSessions > 0 {
		winRate = float64(totalWins) / float64(totalSessions)
	}

	// Win streaks
	currentStreak, bestStreak := calculateWinStreaks(ctx, pool, userID, partyID)

	// Most played game
	var mostPlayedGame *map[string]interface{}
	mpRow := pool.QueryRow(ctx,
		`SELECT g.id, g.name, COUNT(*) AS cnt
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 JOIN games g ON g.id = s.game_id
		 WHERE sp.user_id = $1`+partyFilter+`
		 GROUP BY g.id, g.name ORDER BY cnt DESC LIMIT 1`, args...)
	var mpID, mpName string
	var mpCount int
	if err := mpRow.Scan(&mpID, &mpName, &mpCount); err == nil {
		mp := map[string]interface{}{"id": mpID, "name": mpName, "session_count": mpCount}
		mostPlayedGame = &mp
	}

	// Best win rate game (min 3 sessions)
	var bestWinRateGame *map[string]interface{}
	bwrRow := pool.QueryRow(ctx,
		`SELECT g.id, g.name,
		        COUNT(*) AS sessions,
		        COUNT(*) FILTER (WHERE sp.rank = 1 OR sp.result = 'win') AS wins
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 JOIN games g ON g.id = s.game_id
		 WHERE sp.user_id = $1`+partyFilter+`
		 GROUP BY g.id, g.name
		 HAVING COUNT(*) >= 3
		 ORDER BY (COUNT(*) FILTER (WHERE sp.rank = 1 OR sp.result = 'win'))::float / COUNT(*) DESC
		 LIMIT 1`, args...)
	var bwrID, bwrName string
	var bwrSessions, bwrWins int
	if err := bwrRow.Scan(&bwrID, &bwrName, &bwrSessions, &bwrWins); err == nil {
		bwr := map[string]interface{}{"id": bwrID, "name": bwrName, "win_rate": float64(bwrWins) / float64(bwrSessions)}
		bestWinRateGame = &bwr
	}

	// Nemesis (most losses against)
	nemesis := findNemesis(ctx, pool, userID, partyID)

	// Punching bag (most wins against)
	punchingBag := findPunchingBag(ctx, pool, userID, partyID)

	// Per-game breakdown
	perGame := perGameStats(ctx, pool, userID, partyID)

	// Head-to-head
	h2h := headToHeadAll(ctx, pool, userID, partyID)

	// User info
	var displayName string
	var avatarURL *string
	pool.QueryRow(ctx, "SELECT display_name, avatar_url FROM users WHERE id = $1", userID).Scan(&displayName, &avatarURL)

	result := map[string]interface{}{
		"user":               map[string]interface{}{"id": userID, "display_name": displayName, "avatar_url": avatarURL},
		"total_sessions":     totalSessions,
		"total_wins":         totalWins,
		"win_rate":           winRate,
		"current_streak":     currentStreak,
		"best_streak":        bestStreak,
		"most_played_game":   mostPlayedGame,
		"best_win_rate_game": bestWinRateGame,
		"nemesis":            nemesis,
		"punching_bag":       punchingBag,
		"per_game":           perGame,
		"head_to_head":       h2h,
	}

	return result, nil
}

// CalculateStreaks returns current and best win streaks for a user.
func CalculateStreaks(ctx context.Context, pool *pgxpool.Pool, userID string, partyID *string) (current, best int) {
	return calculateWinStreaks(ctx, pool, userID, partyID)
}

func calculateWinStreaks(ctx context.Context, pool *pgxpool.Pool, userID string, partyID *string) (current, best int) {
	partyFilter := ""
	args := []interface{}{userID}
	if partyID != nil {
		partyFilter = " AND s.party_id = $2"
		args = append(args, *partyID)
	}

	rows, err := pool.Query(ctx,
		`SELECT (sp.rank = 1 OR sp.result = 'win') AS won
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 WHERE sp.user_id = $1`+partyFilter+`
		 ORDER BY s.played_at DESC`, args...)
	if err != nil {
		return 0, 0
	}
	defer rows.Close()

	streak := 0
	currentSet := false
	for rows.Next() {
		var won bool
		rows.Scan(&won)
		if won {
			streak++
		} else {
			if !currentSet {
				current = streak
				currentSet = true
			}
			if streak > best {
				best = streak
			}
			streak = 0
		}
	}
	if !currentSet {
		current = streak
	}
	if streak > best {
		best = streak
	}
	return current, best
}

func findNemesis(ctx context.Context, pool *pgxpool.Pool, userID string, partyID *string) *map[string]interface{} {
	partyFilter := ""
	args := []interface{}{userID}
	if partyID != nil {
		partyFilter = " AND s.party_id = $2"
		args = append(args, *partyID)
	}

	// Find opponent who beat this user the most
	row := pool.QueryRow(ctx,
		`SELECT opp.user_id, u.display_name, COUNT(*) AS losses
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 JOIN session_participants opp ON opp.session_id = s.id AND opp.user_id != $1
		 JOIN users u ON u.id = opp.user_id
		 WHERE sp.user_id = $1`+partyFilter+`
		   AND (sp.rank IS NOT NULL AND opp.rank IS NOT NULL AND opp.rank < sp.rank
		        OR sp.result = 'loss' AND opp.result = 'win')
		 GROUP BY opp.user_id, u.display_name
		 ORDER BY losses DESC LIMIT 1`, args...)

	var oppID, oppName string
	var losses int
	if err := row.Scan(&oppID, &oppName, &losses); err == nil {
		result := map[string]interface{}{
			"id":             oppID,
			"display_name":   oppName,
			"losses_against": losses,
		}
		return &result
	}
	return nil
}

func findPunchingBag(ctx context.Context, pool *pgxpool.Pool, userID string, partyID *string) *map[string]interface{} {
	partyFilter := ""
	args := []interface{}{userID}
	if partyID != nil {
		partyFilter = " AND s.party_id = $2"
		args = append(args, *partyID)
	}

	row := pool.QueryRow(ctx,
		`SELECT opp.user_id, u.display_name, COUNT(*) AS wins
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 JOIN session_participants opp ON opp.session_id = s.id AND opp.user_id != $1
		 JOIN users u ON u.id = opp.user_id
		 WHERE sp.user_id = $1`+partyFilter+`
		   AND (sp.rank IS NOT NULL AND opp.rank IS NOT NULL AND sp.rank < opp.rank
		        OR sp.result = 'win' AND opp.result = 'loss')
		 GROUP BY opp.user_id, u.display_name
		 ORDER BY wins DESC LIMIT 1`, args...)

	var oppID, oppName string
	var wins int
	if err := row.Scan(&oppID, &oppName, &wins); err == nil {
		result := map[string]interface{}{
			"id":           oppID,
			"display_name": oppName,
			"wins_against": wins,
		}
		return &result
	}
	return nil
}

func perGameStats(ctx context.Context, pool *pgxpool.Pool, userID string, partyID *string) []PerGameStat {
	partyFilter := ""
	args := []interface{}{userID}
	if partyID != nil {
		partyFilter = " AND s.party_id = $2"
		args = append(args, *partyID)
	}

	rows, err := pool.Query(ctx,
		`SELECT g.id, g.name, g.cover_image_url,
		        COUNT(DISTINCT sp.session_id) AS sessions,
		        COUNT(DISTINCT sp.session_id) FILTER (WHERE sp.rank = 1 OR sp.result = 'win') AS wins
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 JOIN games g ON g.id = s.game_id
		 WHERE sp.user_id = $1`+partyFilter+`
		 GROUP BY g.id, g.name, g.cover_image_url
		 ORDER BY sessions DESC`, args...)
	if err != nil {
		return []PerGameStat{}
	}
	defer rows.Close()

	stats := []PerGameStat{}
	for rows.Next() {
		var gs PerGameStat
		rows.Scan(&gs.Game.ID, &gs.Game.Name, &gs.Game.CoverImageURL, &gs.Sessions, &gs.Wins)
		if gs.Sessions > 0 {
			gs.WinRate = float64(gs.Wins) / float64(gs.Sessions)
		}
		stats = append(stats, gs)
	}
	return stats
}

func headToHeadAll(ctx context.Context, pool *pgxpool.Pool, userID string, partyID *string) []HeadToHeadEntry {
	partyFilter := ""
	args := []interface{}{userID}
	if partyID != nil {
		partyFilter = " AND s.party_id = $2"
		args = append(args, *partyID)
	}

	// Get all opponents
	rows, err := pool.Query(ctx,
		`SELECT opp.user_id, u.display_name, u.avatar_url,
		        COUNT(DISTINCT s.id) AS sessions_together,
		        COUNT(DISTINCT s.id) FILTER (WHERE (sp.rank = 1 OR sp.result = 'win')) AS this_user_wins,
		        COUNT(DISTINCT s.id) FILTER (WHERE (opp.rank = 1 OR opp.result = 'win')) AS opponent_wins
		 FROM session_participants sp
		 JOIN sessions s ON s.id = sp.session_id
		 JOIN session_participants opp ON opp.session_id = s.id AND opp.user_id != $1
		 JOIN users u ON u.id = opp.user_id
		 WHERE sp.user_id = $1`+partyFilter+`
		 GROUP BY opp.user_id, u.display_name, u.avatar_url
		 ORDER BY sessions_together DESC`, args...)
	if err != nil {
		return []HeadToHeadEntry{}
	}
	defer rows.Close()

	entries := []HeadToHeadEntry{}
	for rows.Next() {
		var e HeadToHeadEntry
		rows.Scan(&e.Opponent.ID, &e.Opponent.DisplayName, &e.Opponent.AvatarURL,
			&e.SessionsTogether, &e.ThisUserWins, &e.OpponentWins)
		entries = append(entries, e)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].SessionsTogether > entries[j].SessionsTogether
	})

	return entries
}
