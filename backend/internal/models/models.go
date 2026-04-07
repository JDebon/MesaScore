package models

import (
	"time"
)

type User struct {
	ID                 string     `json:"id"`
	Username           string     `json:"username"`
	DisplayName        string     `json:"display_name"`
	Email              string     `json:"email,omitempty"`
	PasswordHash       string     `json:"-"`
	EmailVerified      bool       `json:"-"`
	VerificationToken  *string    `json:"-"`
	VerificationSentAt *time.Time `json:"-"`
	AvatarURL          *string    `json:"avatar_url"`
	CreatedAt          time.Time  `json:"created_at"`
}

type Party struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	AdminUserID string    `json:"admin_user_id,omitempty"`
	InviteCode  string    `json:"invite_code,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type PartyMember struct {
	ID       string    `json:"id"`
	PartyID  string    `json:"party_id"`
	UserID   string    `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type PartyInvite struct {
	ID              string    `json:"id"`
	PartyID         string    `json:"party_id"`
	InvitedUserID   string    `json:"invited_user_id"`
	InvitedByUserID string    `json:"invited_by_user_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type Game struct {
	ID             string     `json:"id"`
	BggID          *int       `json:"bgg_id"`
	Name           string     `json:"name"`
	Description    *string    `json:"description"`
	CoverImageURL  *string    `json:"cover_image_url"`
	MinPlayers     *int       `json:"min_players"`
	MaxPlayers     *int       `json:"max_players"`
	BggRating      *float64   `json:"bgg_rating"`
	BggFetchedAt   *time.Time `json:"bgg_fetched_at"`
	AddedByUserID  string     `json:"added_by_user_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type UserGame struct {
	ID      string    `json:"id"`
	UserID  string    `json:"user_id"`
	GameID  string    `json:"game_id"`
	AddedAt time.Time `json:"added_at"`
}

type Session struct {
	ID               string    `json:"id"`
	PartyID          string    `json:"party_id"`
	GameID           string    `json:"game_id"`
	SessionType      string    `json:"session_type"`
	PlayedAt         time.Time `json:"played_at"`
	DurationMinutes  *int      `json:"duration_minutes"`
	BroughtByUserID  *string   `json:"brought_by_user_id"`
	Notes            *string   `json:"notes"`
	CreatedBy        string    `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
}

type SessionParticipant struct {
	ID        string   `json:"id"`
	SessionID string   `json:"session_id"`
	UserID    string   `json:"user_id"`
	TeamName  *string  `json:"team_name"`
	Rank      *int     `json:"rank"`
	Score     *float64 `json:"score"`
	Result    *string  `json:"result"`
}
