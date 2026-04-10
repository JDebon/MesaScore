package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"mesascore/internal/bgg"
	"mesascore/internal/db"
	"mesascore/internal/email"
	"mesascore/internal/handlers"
	mw "mesascore/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Config
	databaseURL := envOrDefault("DATABASE_URL", "postgres://mesascore:changeme@localhost:5432/mesascore?sslmode=disable")
	jwtSecret := envOrDefault("JWT_SECRET", "dev-secret")
	baseURL := envOrDefault("BASE_URL", "http://localhost")
	allowedOrigin := envOrDefault("ALLOWED_ORIGIN", "http://localhost:5173")
	port := envOrDefault("PORT", "8080")
	smtpHost := envOrDefault("SMTP_HOST", "localhost")
	smtpPort := envOrDefault("SMTP_PORT", "1025")
	smtpUser := envOrDefault("SMTP_USER", "")
	smtpPass := envOrDefault("SMTP_PASS", "")
	smtpFrom := envOrDefault("SMTP_FROM", "MesaScore <noreply@mesascore.local>")
	bggAPIToken := envOrDefault("BGG_API_TOKEN", "")

	bgg.Init(bggAPIToken)

	ctx := context.Background()

	// Migrations
	if err := db.RunMigrations(databaseURL, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Database pool
	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Services
	mailer := email.NewSender(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom, baseURL)

	// Handlers
	authHandler := handlers.NewAuthHandler(pool, jwtSecret, mailer)
	userHandler := handlers.NewUserHandler(pool)
	partyHandler := handlers.NewPartyHandler(pool)
	collectionHandler := handlers.NewCollectionHandler(pool)
	gameHandler := handlers.NewGameHandler(pool)
	sessionHandler := handlers.NewSessionHandler(pool)
	statsHandler := handlers.NewStatsHandler(pool)
	dashboardHandler := handlers.NewDashboardHandler(pool)
	bggHandler := handlers.NewBGGHandler()

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{allowedOrigin},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		ExposedHeaders: []string{"X-New-Token"},
		MaxAge:         300,
	}))

	// Public routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Get("/api/auth/verify-email", authHandler.VerifyEmail)
	r.Post("/api/auth/resend-verification", authHandler.ResendVerification)
	r.Get("/api/auth/check-username", authHandler.CheckUsername)
	r.Get("/api/parties/join/{inviteCode}", partyHandler.JoinPreview)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(mw.AuthMiddleware(jwtSecret))

		// User endpoints
		r.Get("/api/users/me", userHandler.GetMe)
		r.Patch("/api/users/me", userHandler.UpdateMe)
		r.Get("/api/users/me/dashboard", dashboardHandler.UserDashboard)
		r.Get("/api/users/me/invites", partyHandler.PendingInvites)
		r.Post("/api/users/me/collection", collectionHandler.Add)
		r.Delete("/api/users/me/collection/{gameId}", collectionHandler.Remove)

		r.Get("/api/users/search", userHandler.Search)
		r.Get("/api/users/{id}", userHandler.Get)
		r.Get("/api/users/{id}/stats", statsHandler.UserGlobal)
		r.Get("/api/users/{id}/collection", collectionHandler.List)

		// Party creation and join
		r.Post("/api/parties", partyHandler.Create)
		r.Post("/api/parties/join/{inviteCode}", partyHandler.Join)

		// Invite accept/decline — invited user is not yet a member
		r.Post("/api/parties/{partyId}/invites/{inviteId}/accept", partyHandler.AcceptInvite)
		r.Post("/api/parties/{partyId}/invites/{inviteId}/decline", partyHandler.DeclineInvite)

		// Game catalog (global)
		r.Get("/api/games", gameHandler.List)
		r.Get("/api/games/{id}", gameHandler.Get)
		r.Post("/api/games", gameHandler.Create)
		r.Patch("/api/games/{id}", gameHandler.Update)
		r.Post("/api/games/{id}/bgg-refresh", gameHandler.BGGRefresh)
		r.Get("/api/bgg/search", bggHandler.Search)
		r.Get("/api/bgg/thing", bggHandler.Thing)

		// Party-scoped (member access)
		r.Group(func(r chi.Router) {
			r.Use(mw.PartyMemberMiddleware(pool))

			r.Get("/api/parties/{partyId}", partyHandler.Get)
			r.Get("/api/parties/{partyId}/dashboard", dashboardHandler.PartyDashboard)
			r.Get("/api/parties/{partyId}/members", partyHandler.Members)
			r.Get("/api/parties/{partyId}/available-games", gameHandler.AvailableForParty)
			r.Get("/api/parties/{partyId}/sessions", sessionHandler.List)
			r.Get("/api/parties/{partyId}/sessions/{sessionId}", sessionHandler.Get)
			r.Post("/api/parties/{partyId}/sessions", sessionHandler.Create)
			r.Get("/api/parties/{partyId}/stats/leaderboard", statsHandler.PartyLeaderboard)
			r.Get("/api/parties/{partyId}/stats/games/{gameId}", statsHandler.PartyGame)
			r.Get("/api/parties/{partyId}/stats/activity", statsHandler.PartyActivity)
			r.Get("/api/parties/{partyId}/users/{userId}/stats", statsHandler.UserInParty)

			r.Post("/api/parties/{partyId}/leave", partyHandler.Leave)
		})

		// Party-scoped (admin access)
		r.Group(func(r chi.Router) {
			r.Use(mw.PartyAdminMiddleware(pool))

			r.Patch("/api/parties/{partyId}", partyHandler.Update)
			r.Post("/api/parties/{partyId}/regenerate-invite", partyHandler.RegenerateInvite)
			r.Post("/api/parties/{partyId}/transfer-admin", partyHandler.TransferAdmin)
			r.Delete("/api/parties/{partyId}/members/{userId}", partyHandler.RemoveMember)
			r.Post("/api/parties/{partyId}/invites", partyHandler.SendInvite)
			r.Patch("/api/parties/{partyId}/sessions/{sessionId}", sessionHandler.Update)
			r.Delete("/api/parties/{partyId}/sessions/{sessionId}", sessionHandler.Delete)
		})
	})

	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
