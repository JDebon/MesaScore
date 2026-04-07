package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func generateInviteCode() string {
	b := make([]byte, 24)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:32]
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://mesascore:changeme@localhost:5432/mesascore?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Clear existing seed data (reverse order of dependencies)
	tables := []string{
		"session_participants", "sessions", "user_games",
		"games", "party_invites", "party_members", "parties", "users",
	}
	for _, t := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", t)); err != nil {
			log.Fatalf("failed to clear %s: %v", t, err)
		}
	}

	log.Println("Cleared existing data")

	// --- Users (T1.11: 6 users, all verified) ---
	pw := hashPassword("password123")
	users := []struct {
		id, username, displayName, email string
	}{
		{"a0000000-0000-0000-0000-000000000001", "alice", "Alice García", "alice@example.com"},
		{"a0000000-0000-0000-0000-000000000002", "bob", "Bob Martínez", "bob@example.com"},
		{"a0000000-0000-0000-0000-000000000003", "carol", "Carol López", "carol@example.com"},
		{"a0000000-0000-0000-0000-000000000004", "david", "David Ruiz", "david@example.com"},
		{"a0000000-0000-0000-0000-000000000005", "eva", "Eva Fernández", "eva@example.com"},
		{"a0000000-0000-0000-0000-000000000006", "frank", "Frank Torres", "frank@example.com"},
	}

	for _, u := range users {
		_, err := db.ExecContext(ctx,
			`INSERT INTO users (id, username, display_name, email, password_hash, email_verified, created_at)
			 VALUES ($1, $2, $3, $4, $5, TRUE, NOW())`,
			u.id, u.username, u.displayName, u.email, pw,
		)
		if err != nil {
			log.Fatalf("insert user %s: %v", u.username, err)
		}
	}
	log.Println("Created 6 users")

	// --- Parties (2 parties, alice admin of party1, bob admin of party2) ---
	parties := []struct {
		id, name, description, adminID string
	}{
		{"b0000000-0000-0000-0000-000000000001", "Friday Night Games", "Weekly board game meetup", users[0].id},
		{"b0000000-0000-0000-0000-000000000002", "Weekend Warriors", "Saturday afternoon sessions", users[1].id},
	}

	for _, p := range parties {
		_, err := db.ExecContext(ctx,
			`INSERT INTO parties (id, name, description, admin_user_id, invite_code, created_at)
			 VALUES ($1, $2, $3, $4, $5, NOW())`,
			p.id, p.name, p.description, p.adminID, generateInviteCode(),
		)
		if err != nil {
			log.Fatalf("insert party %s: %v", p.name, err)
		}
	}
	log.Println("Created 2 parties")

	// --- Party Members ---
	// Party 1 (Friday Night Games): alice(admin), bob, carol, david, eva — 5 members
	// Party 2 (Weekend Warriors): bob(admin), carol, david, frank — 4 members
	// Note: bob belongs to both parties
	memberships := []struct{ partyID, userID string }{
		{parties[0].id, users[0].id}, // alice in party1
		{parties[0].id, users[1].id}, // bob in party1
		{parties[0].id, users[2].id}, // carol in party1
		{parties[0].id, users[3].id}, // david in party1
		{parties[0].id, users[4].id}, // eva in party1
		{parties[1].id, users[1].id}, // bob in party2
		{parties[1].id, users[2].id}, // carol in party2
		{parties[1].id, users[3].id}, // david in party2
		{parties[1].id, users[5].id}, // frank in party2
	}

	for _, m := range memberships {
		_, err := db.ExecContext(ctx,
			`INSERT INTO party_members (id, party_id, user_id, joined_at) VALUES (gen_random_uuid(), $1, $2, NOW())`,
			m.partyID, m.userID,
		)
		if err != nil {
			log.Fatalf("insert member: %v", err)
		}
	}
	log.Println("Created party memberships")

	// --- Games (15 games in catalog) ---
	games := []struct {
		id, name    string
		bggID       int
		minP, maxP  int
		addedByID   string
	}{
		{"c0000000-0000-0000-0000-000000000001", "Catan", 13, 3, 4, users[0].id},
		{"c0000000-0000-0000-0000-000000000002", "Wingspan", 266192, 1, 5, users[0].id},
		{"c0000000-0000-0000-0000-000000000003", "Ticket to Ride", 9209, 2, 5, users[0].id},
		{"c0000000-0000-0000-0000-000000000004", "Pandemic", 30549, 2, 4, users[1].id},
		{"c0000000-0000-0000-0000-000000000005", "Azul", 230802, 2, 4, users[1].id},
		{"c0000000-0000-0000-0000-000000000006", "7 Wonders", 68448, 2, 7, users[1].id},
		{"c0000000-0000-0000-0000-000000000007", "Codenames", 178900, 2, 8, users[2].id},
		{"c0000000-0000-0000-0000-000000000008", "Terraforming Mars", 167791, 1, 5, users[2].id},
		{"c0000000-0000-0000-0000-000000000009", "Spirit Island", 162886, 1, 4, users[3].id},
		{"c0000000-0000-0000-0000-000000000010", "Scythe", 169786, 1, 5, users[3].id},
		{"c0000000-0000-0000-0000-000000000011", "Root", 237182, 2, 4, users[4].id},
		{"c0000000-0000-0000-0000-000000000012", "Everdell", 199792, 1, 4, users[4].id},
		{"c0000000-0000-0000-0000-000000000013", "Gloomhaven", 174430, 1, 4, users[5].id},
		{"c0000000-0000-0000-0000-000000000014", "Brass: Birmingham", 224517, 2, 4, users[5].id},
		{"c0000000-0000-0000-0000-000000000015", "Ark Nova", 342942, 1, 4, users[0].id},
	}

	for _, g := range games {
		_, err := db.ExecContext(ctx,
			`INSERT INTO games (id, bgg_id, name, min_players, max_players, added_by_user_id, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
			g.id, g.bggID, g.name, g.minP, g.maxP, g.addedByID,
		)
		if err != nil {
			log.Fatalf("insert game %s: %v", g.name, err)
		}
	}
	log.Println("Created 15 games")

	// --- User Games (distribute across collections) ---
	// Each user who added a game owns it, plus some extras
	userGames := []struct{ userID, gameID string }{
		// Alice owns: Catan, Wingspan, Ticket to Ride, Ark Nova
		{users[0].id, games[0].id}, {users[0].id, games[1].id}, {users[0].id, games[2].id}, {users[0].id, games[14].id},
		// Bob owns: Pandemic, Azul, 7 Wonders, Catan
		{users[1].id, games[3].id}, {users[1].id, games[4].id}, {users[1].id, games[5].id}, {users[1].id, games[0].id},
		// Carol owns: Codenames, Terraforming Mars, Wingspan
		{users[2].id, games[6].id}, {users[2].id, games[7].id}, {users[2].id, games[1].id},
		// David owns: Spirit Island, Scythe, Azul
		{users[3].id, games[8].id}, {users[3].id, games[9].id}, {users[3].id, games[4].id},
		// Eva owns: Root, Everdell
		{users[4].id, games[10].id}, {users[4].id, games[11].id},
		// Frank owns: Gloomhaven, Brass: Birmingham, Scythe
		{users[5].id, games[12].id}, {users[5].id, games[13].id}, {users[5].id, games[9].id},
	}

	for _, ug := range userGames {
		_, err := db.ExecContext(ctx,
			`INSERT INTO user_games (id, user_id, game_id, added_at) VALUES (gen_random_uuid(), $1, $2, NOW())`,
			ug.userID, ug.gameID,
		)
		if err != nil {
			log.Fatalf("insert user_game: %v", err)
		}
	}
	log.Println("Created user game collections")

	// --- Sessions (5 per party, mix of session types) ---
	baseTime := time.Now().AddDate(0, -1, 0) // Start from a month ago

	type sessionDef struct {
		id, partyID, gameID, sessionType, createdBy, broughtBy string
		dayOffset                                               int
		duration                                                int
	}

	sessionDefs := []sessionDef{
		// Party 1 sessions (created by alice, the admin)
		{"d0000000-0000-0000-0000-000000000001", parties[0].id, games[0].id, "competitive", users[0].id, users[0].id, 0, 90},
		{"d0000000-0000-0000-0000-000000000002", parties[0].id, games[6].id, "team", users[0].id, users[2].id, 7, 45},
		{"d0000000-0000-0000-0000-000000000003", parties[0].id, games[3].id, "cooperative", users[0].id, users[1].id, 14, 60},
		{"d0000000-0000-0000-0000-000000000004", parties[0].id, games[1].id, "score", users[0].id, users[0].id, 21, 75},
		{"d0000000-0000-0000-0000-000000000005", parties[0].id, games[4].id, "competitive", users[0].id, users[1].id, 28, 40},
		// Party 2 sessions (created by bob, the admin)
		{"d0000000-0000-0000-0000-000000000006", parties[1].id, games[9].id, "competitive", users[1].id, users[5].id, 2, 120},
		{"d0000000-0000-0000-0000-000000000007", parties[1].id, games[8].id, "cooperative", users[1].id, users[3].id, 9, 90},
		{"d0000000-0000-0000-0000-000000000008", parties[1].id, games[13].id, "score", users[1].id, users[5].id, 16, 100},
		{"d0000000-0000-0000-0000-000000000009", parties[1].id, games[6].id, "team", users[1].id, users[2].id, 23, 30},
		{"d0000000-0000-0000-0000-000000000010", parties[1].id, games[4].id, "competitive", users[1].id, users[3].id, 28, 50},
	}

	for _, s := range sessionDefs {
		playedAt := baseTime.AddDate(0, 0, s.dayOffset)
		_, err := db.ExecContext(ctx,
			`INSERT INTO sessions (id, party_id, game_id, session_type, played_at, duration_minutes, brought_by_user_id, created_by, created_at)
			 VALUES ($1, $2, $3, $4::session_type, $5, $6, $7, $8, NOW())`,
			s.id, s.partyID, s.gameID, s.sessionType, playedAt, s.duration, s.broughtBy, s.createdBy,
		)
		if err != nil {
			log.Fatalf("insert session %s: %v", s.id, err)
		}
	}
	log.Println("Created 10 sessions")

	// --- Session Participants ---
	type participantDef struct {
		sessionID, userID string
		teamName          *string
		rank              *int
		score             *float64
		result            *string
	}

	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }
	floatPtr := func(f float64) *float64 { return &f }

	participants := []participantDef{
		// Session 1: Catan, competitive (party1: alice, bob, carol, david)
		{sessionDefs[0].id, users[0].id, nil, intPtr(1), nil, nil},
		{sessionDefs[0].id, users[1].id, nil, intPtr(2), nil, nil},
		{sessionDefs[0].id, users[2].id, nil, intPtr(3), nil, nil},
		{sessionDefs[0].id, users[3].id, nil, intPtr(4), nil, nil},

		// Session 2: Codenames, team (party1: alice+bob vs carol+david+eva)
		{sessionDefs[1].id, users[0].id, strPtr("Red"), intPtr(1), nil, nil},
		{sessionDefs[1].id, users[1].id, strPtr("Red"), intPtr(1), nil, nil},
		{sessionDefs[1].id, users[2].id, strPtr("Blue"), intPtr(2), nil, nil},
		{sessionDefs[1].id, users[3].id, strPtr("Blue"), intPtr(2), nil, nil},
		{sessionDefs[1].id, users[4].id, strPtr("Blue"), intPtr(2), nil, nil},

		// Session 3: Pandemic, cooperative (party1: alice, bob, carol)
		{sessionDefs[2].id, users[0].id, nil, nil, nil, strPtr("win")},
		{sessionDefs[2].id, users[1].id, nil, nil, nil, strPtr("win")},
		{sessionDefs[2].id, users[2].id, nil, nil, nil, strPtr("win")},

		// Session 4: Wingspan, score (party1: alice, bob, eva)
		{sessionDefs[3].id, users[0].id, nil, nil, floatPtr(82.5), nil},
		{sessionDefs[3].id, users[1].id, nil, nil, floatPtr(71.0), nil},
		{sessionDefs[3].id, users[4].id, nil, nil, floatPtr(94.0), nil},

		// Session 5: Azul, competitive (party1: alice, carol, david)
		{sessionDefs[4].id, users[0].id, nil, intPtr(2), nil, nil},
		{sessionDefs[4].id, users[2].id, nil, intPtr(1), nil, nil},
		{sessionDefs[4].id, users[3].id, nil, intPtr(3), nil, nil},

		// Session 6: Scythe, competitive (party2: bob, carol, david, frank)
		{sessionDefs[5].id, users[1].id, nil, intPtr(1), nil, nil},
		{sessionDefs[5].id, users[2].id, nil, intPtr(3), nil, nil},
		{sessionDefs[5].id, users[3].id, nil, intPtr(2), nil, nil},
		{sessionDefs[5].id, users[5].id, nil, intPtr(4), nil, nil},

		// Session 7: Spirit Island, cooperative (party2: bob, david, frank)
		{sessionDefs[6].id, users[1].id, nil, nil, nil, strPtr("loss")},
		{sessionDefs[6].id, users[3].id, nil, nil, nil, strPtr("loss")},
		{sessionDefs[6].id, users[5].id, nil, nil, nil, strPtr("loss")},

		// Session 8: Brass: Birmingham, score (party2: bob, carol, frank)
		{sessionDefs[7].id, users[1].id, nil, nil, floatPtr(145.0), nil},
		{sessionDefs[7].id, users[2].id, nil, nil, floatPtr(132.0), nil},
		{sessionDefs[7].id, users[5].id, nil, nil, floatPtr(158.0), nil},

		// Session 9: Codenames, team (party2: bob+frank vs carol+david)
		{sessionDefs[8].id, users[1].id, strPtr("Alpha"), intPtr(2), nil, nil},
		{sessionDefs[8].id, users[5].id, strPtr("Alpha"), intPtr(2), nil, nil},
		{sessionDefs[8].id, users[2].id, strPtr("Beta"), intPtr(1), nil, nil},
		{sessionDefs[8].id, users[3].id, strPtr("Beta"), intPtr(1), nil, nil},

		// Session 10: Azul, competitive (party2: bob, david, frank)
		{sessionDefs[9].id, users[1].id, nil, intPtr(2), nil, nil},
		{sessionDefs[9].id, users[3].id, nil, intPtr(1), nil, nil},
		{sessionDefs[9].id, users[5].id, nil, intPtr(3), nil, nil},
	}

	for _, p := range participants {
		_, err := db.ExecContext(ctx,
			`INSERT INTO session_participants (id, session_id, user_id, team_name, rank, score, result)
			 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6::participant_result)`,
			p.sessionID, p.userID, p.teamName, p.rank, p.score, p.result,
		)
		if err != nil {
			log.Fatalf("insert participant: %v", err)
		}
	}
	log.Println("Created session participants")

	log.Println("Seed complete!")
}
