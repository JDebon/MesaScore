package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ParticipantInput struct {
	UserID   string   `json:"user_id"`
	TeamName *string  `json:"team_name"`
	Rank     *int     `json:"rank"`
	Score    *float64 `json:"score"`
	Result   *string  `json:"result"`
}

type SessionInput struct {
	GameID          string             `json:"game_id"`
	SessionType     string             `json:"session_type"`
	PlayedAt        time.Time          `json:"played_at"`
	DurationMinutes *int               `json:"duration_minutes"`
	Notes           *string            `json:"notes"`
	BroughtByUserID *string            `json:"brought_by_user_id"`
	Participants    []ParticipantInput `json:"participants"`
}

func ValidateSession(ctx context.Context, pool *pgxpool.Pool, partyID string, input SessionInput) map[string]string {
	errors := map[string]string{}

	// Session type
	validTypes := map[string]bool{"competitive": true, "team": true, "cooperative": true, "score": true}
	if !validTypes[input.SessionType] {
		errors["session_type"] = "must be competitive, team, cooperative, or score"
	}

	// Date not in future
	if input.PlayedAt.After(time.Now().Add(1 * time.Hour)) {
		errors["played_at"] = "cannot be in the future"
	}

	// Min 2 participants
	if len(input.Participants) < 2 {
		errors["participants"] = "at least 2 participants required"
	}

	// Game exists
	var gameExists bool
	pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM games WHERE id = $1)", input.GameID).Scan(&gameExists)
	if !gameExists {
		errors["game_id"] = "game not found"
	}

	// All participants must be party members
	for i, p := range input.Participants {
		var isMember bool
		pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM party_members WHERE party_id = $1 AND user_id = $2)",
			partyID, p.UserID).Scan(&isMember)
		if !isMember {
			errors[fmt.Sprintf("participants[%d].user_id", i)] = "not a party member"
		}
	}

	// brought_by must be a member who owns the game
	if input.BroughtByUserID != nil {
		var isMember bool
		pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM party_members WHERE party_id = $1 AND user_id = $2)",
			partyID, *input.BroughtByUserID).Scan(&isMember)
		if !isMember {
			errors["brought_by_user_id"] = "not a party member"
		} else {
			var ownsGame bool
			pool.QueryRow(ctx,
				"SELECT EXISTS(SELECT 1 FROM user_games WHERE user_id = $1 AND game_id = $2)",
				*input.BroughtByUserID, input.GameID).Scan(&ownsGame)
			if !ownsGame {
				errors["brought_by_user_id"] = "user does not own this game"
			}
		}
	}

	// Validate participant fields based on session type
	if len(errors) == 0 {
		switch input.SessionType {
		case "competitive":
			for i, p := range input.Participants {
				if p.Rank == nil {
					errors[fmt.Sprintf("participants[%d].rank", i)] = "required for competitive sessions"
				}
			}
		case "team":
			for i, p := range input.Participants {
				if p.TeamName == nil {
					errors[fmt.Sprintf("participants[%d].team_name", i)] = "required for team sessions"
				}
				if p.Rank == nil {
					errors[fmt.Sprintf("participants[%d].rank", i)] = "required for team sessions"
				}
			}
		case "cooperative":
			for i, p := range input.Participants {
				if p.Result == nil {
					errors[fmt.Sprintf("participants[%d].result", i)] = "required for cooperative sessions"
				} else if *p.Result != "win" && *p.Result != "loss" {
					errors[fmt.Sprintf("participants[%d].result", i)] = "must be win or loss"
				}
			}
		case "score":
			for i, p := range input.Participants {
				if p.Score == nil {
					errors[fmt.Sprintf("participants[%d].score", i)] = "required for score sessions"
				}
			}
		}
	}

	return errors
}
