# API — Technical Tasks

## Stack
- **Language**: Go 1.23+
- **Router**: `go-chi/chi` v5
- **JWT**: `golang-jwt/jwt` v5
- **Password hashing**: `golang.org/x/crypto/bcrypt`
- **Database driver**: `jackc/pgx` v5
- **Migrations**: `golang-migrate/migrate` v4
- **BGG**: BGG XML API 2 (HTTP GET, `encoding/xml`)
- **Email**: `gopkg.in/gomail.v2`

---

## Tasks

### T9.1 — Go module and dependency setup

```bash
go mod init mesascore
go get github.com/go-chi/chi/v5
go get github.com/go-chi/chi/v5/middleware
go get github.com/go-chi/cors
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
go get github.com/jackc/pgx/v5
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file
go get github.com/google/uuid
go get gopkg.in/gomail.v2
```

---

### T9.2 — Application entry point

`cmd/server/main.go`:
- Load config from env: `DATABASE_URL`, `JWT_SECRET`, `BASE_URL`, `ALLOWED_ORIGIN`, `PORT`, `SMTP_*`.
- Run migrations on startup via `golang-migrate`.
- Initialize `pgx.Pool`.
- Build Chi router.
- Start HTTP server.

---

### T9.3 — Router structure

```go
r := chi.NewRouter()
r.Use(middleware.Logger, middleware.Recoverer, corsHandler)

// Public
r.Post("/api/auth/register", authHandler.Register)
r.Post("/api/auth/login", authHandler.Login)
r.Get("/api/auth/verify-email", authHandler.VerifyEmail)
r.Post("/api/auth/resend-verification", authHandler.ResendVerification)
r.Get("/api/auth/check-username", authHandler.CheckUsername)
r.Get("/api/parties/join/{inviteCode}", partyHandler.JoinPreview)

// Authenticated
r.Group(func(r chi.Router) {
    r.Use(AuthMiddleware)

    r.Get("/api/users/me", userHandler.GetMe)
    r.Patch("/api/users/me", userHandler.UpdateMe)
    r.Get("/api/users/me/dashboard", userHandler.Dashboard)
    r.Get("/api/users/me/invites", userHandler.PendingInvites)
    r.Post("/api/users/me/collection", collectionHandler.Add)
    r.Delete("/api/users/me/collection/{gameId}", collectionHandler.Remove)

    r.Get("/api/users/search", userHandler.Search)
    r.Get("/api/users/{id}", userHandler.Get)
    r.Get("/api/users/{id}/stats", statsHandler.UserGlobal)
    r.Get("/api/users/{id}/collection", collectionHandler.List)

    r.Post("/api/parties", partyHandler.Create)
    r.Post("/api/parties/join/{inviteCode}", partyHandler.Join)

    // Party-scoped (member access)
    r.Group(func(r chi.Router) {
        r.Use(PartyMemberMiddleware)

        r.Get("/api/parties/{partyId}", partyHandler.Get)
        r.Get("/api/parties/{partyId}/dashboard", partyHandler.Dashboard)
        r.Get("/api/parties/{partyId}/members", partyHandler.Members)
        r.Get("/api/parties/{partyId}/available-games", gameHandler.AvailableForParty)
        r.Get("/api/parties/{partyId}/sessions", sessionHandler.List)
        r.Get("/api/parties/{partyId}/sessions/{sessionId}", sessionHandler.Get)
        r.Get("/api/parties/{partyId}/stats/leaderboard", statsHandler.PartyLeaderboard)
        r.Get("/api/parties/{partyId}/stats/games/{gameId}", statsHandler.PartyGame)
        r.Get("/api/parties/{partyId}/stats/activity", statsHandler.PartyActivity)
        r.Get("/api/parties/{partyId}/users/{userId}/stats", statsHandler.UserInParty)

        // Invite response (invited user only — checked inside handler)
        r.Post("/api/parties/{partyId}/invites/{inviteId}/accept", partyHandler.AcceptInvite)
        r.Post("/api/parties/{partyId}/invites/{inviteId}/decline", partyHandler.DeclineInvite)

        r.Post("/api/parties/{partyId}/leave", partyHandler.Leave)
    })

    // Party-scoped (admin access)
    r.Group(func(r chi.Router) {
        r.Use(PartyAdminMiddleware)

        r.Patch("/api/parties/{partyId}", partyHandler.Update)
        r.Post("/api/parties/{partyId}/regenerate-invite", partyHandler.RegenerateInvite)
        r.Post("/api/parties/{partyId}/transfer-admin", partyHandler.TransferAdmin)
        r.Delete("/api/parties/{partyId}/members/{userId}", partyHandler.RemoveMember)
        r.Post("/api/parties/{partyId}/invites", partyHandler.SendInvite)
        r.Post("/api/parties/{partyId}/sessions", sessionHandler.Create)
        r.Patch("/api/parties/{partyId}/sessions/{sessionId}", sessionHandler.Update)
        r.Delete("/api/parties/{partyId}/sessions/{sessionId}", sessionHandler.Delete)
    })

    r.Get("/api/games", gameHandler.List)
    r.Get("/api/games/{id}", gameHandler.Get)
    r.Post("/api/games", gameHandler.Create)
    r.Patch("/api/games/{id}", gameHandler.Update)
    r.Post("/api/games/{id}/bgg-refresh", gameHandler.BGGRefresh)
    r.Get("/api/bgg/search", bggHandler.Search)
})
```

---

### T9.4 — Middleware

**`AuthMiddleware`**
- Extracts Bearer token from `Authorization` header.
- Validates JWT signature and expiry.
- Attaches `userID` to request context.
- Returns `401` on failure.
- Checks token refresh window, issues `X-New-Token` if needed.

**`PartyMemberMiddleware`**
- Reads `partyId` from URL param.
- Queries `party_members` for `(party_id, user_id)`.
- Returns `403` if not a member, `404` if party not found.
- Attaches party to context for downstream handlers.

**`PartyAdminMiddleware`**
- Reads `partyId` from URL param.
- Queries `parties.admin_user_id` and compares to context user.
- Returns `403` if not the admin.

---

### T9.5 — Email service

`internal/email/sender.go`:

```go
type Sender struct {
    dialer *gomail.Dialer
    from   string
    baseURL string
}

func (s *Sender) SendVerificationEmail(toEmail, token string) error {
    verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)
    m := gomail.NewMessage()
    m.SetHeader("From", s.from)
    m.SetHeader("To", toEmail)
    m.SetHeader("Subject", "Confirm your MesaScore account")
    m.SetBody("text/plain", fmt.Sprintf("Click to verify your account: %s\n\nThis link expires in 24 hours.", verifyURL))
    m.AddAlternative("text/html", buildVerificationHTML(verifyURL))
    return s.dialer.DialAndSend(m)
}
```

For development: configure `SMTP_HOST=localhost`, `SMTP_PORT=1025` and use `mailpit` or `mailhog` as a local SMTP catcher (add as a Docker Compose service).

---

### T9.6 — Token generation helpers

`internal/auth/tokens.go`:

```go
// Verification tokens: UUID v4
func GenerateVerificationToken() string {
    return uuid.NewString()
}

// Invite codes: 32-char URL-safe base64
func GenerateInviteCode() string {
    b := make([]byte, 24)
    _, _ = rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)[:32]
}

// JWT
func IssueJWT(userID uuid.UUID, secret string, expiry time.Duration) (string, error) { ... }
func ParseJWT(tokenStr, secret string) (*Claims, error) { ... }
```

---

### T9.7 — BGG client

`internal/bgg/client.go` — unchanged from original spec, except the auth restriction is removed (any user can call `/api/bgg/search`).

Key note: the `/thing` endpoint may return HTTP 202 on first call. Retry once after 2 seconds.

---

### T9.8 — Session validation

`internal/validator/session.go` — identical logic to original spec, with one additional rule:

```go
// All participant user_ids must be current members of the session's party
func validateParticipantsAreMembersOfParty(ctx, db, partyID, participantIDs) error { ... }
```

---

### T9.9 — Stats query implementation

`internal/stats/` — all functions accept optional `partyID *uuid.UUID`:
- When `partyID != nil`: filter `sessions.party_id = *partyID`.
- When `partyID == nil`: no party filter (global user stats).

```go
func Leaderboard(ctx context.Context, db *pgx.Pool, partyID uuid.UUID) ([]LeaderboardEntry, error)
func GameStats(ctx context.Context, db *pgx.Pool, partyID, gameID uuid.UUID) (GameStatsResult, error)
func UserStats(ctx context.Context, db *pgx.Pool, userID uuid.UUID, partyID *uuid.UUID) (UserStatsResult, error)
func ActivityPerMonth(ctx context.Context, db *pgx.Pool, partyID uuid.UUID) ([]MonthCount, error)
func WinStreak(ctx context.Context, db *pgx.Pool, userID uuid.UUID, partyID *uuid.UUID) (current, best int, error)
func Nemesis(ctx context.Context, db *pgx.Pool, userID uuid.UUID, partyID *uuid.UUID) (*PlayerRef, error)
func PunchingBag(ctx context.Context, db *pgx.Pool, userID uuid.UUID, partyID *uuid.UUID) (*PlayerRef, error)
func HeadToHead(ctx context.Context, db *pgx.Pool, userID, opponentID uuid.UUID, partyID *uuid.UUID) (HeadToHeadResult, error)
```

Win streak calculation: fetch ordered sessions in Go, iterate — do not use SQL window functions.

---

### T9.10 — CORS configuration

```go
cors.Handler(cors.Options{
    AllowedOrigins: []string{os.Getenv("ALLOWED_ORIGIN")},
    AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{"Authorization", "Content-Type"},
    ExposedHeaders: []string{"X-New-Token"},
    MaxAge:         300,
})
```

---

### T9.11 — Dockerfiles

**Backend Dockerfile** — unchanged from original spec.

**Frontend Dockerfile** — unchanged from original spec (SvelteKit Node adapter).

---

### T9.12 — Docker Compose (updated)

```yaml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: mesascore
      POSTGRES_USER: mesascore
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  mailpit:
    image: axllent/mailpit
    ports:
      - "8025:8025"   # web UI for dev — not exposed in prod
    profiles: ["dev"] # only runs in dev profile

  backend:
    build: ./backend
    environment:
      DATABASE_URL: postgres://mesascore:${POSTGRES_PASSWORD}@postgres:5432/mesascore
      JWT_SECRET: ${JWT_SECRET}
      BASE_URL: ${BASE_URL}
      SMTP_HOST: ${SMTP_HOST}
      SMTP_PORT: ${SMTP_PORT}
      SMTP_USER: ${SMTP_USER}
      SMTP_PASS: ${SMTP_PASS}
      SMTP_FROM: ${SMTP_FROM}
      ALLOWED_ORIGIN: ${ALLOWED_ORIGIN}
    depends_on:
      - postgres

  frontend:
    build: ./frontend
    environment:
      PUBLIC_API_URL: ${BASE_URL}/api

  caddy:
    image: caddy:2
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - backend
      - frontend

volumes:
  postgres_data:
  caddy_data:
```

---

### T9.13 — Implementation order

1. DB migrations (all tables)
2. Register + verify email + login endpoints + email service
3. JWT middleware + party middlewares
4. User profile + collection endpoints
5. Party CRUD + join via invite link
6. Invite system (username search + accept/decline)
7. Member management (remove, leave, transfer)
8. Game catalog + BGG proxy
9. Session CRUD + validation (party-scoped)
10. Stats endpoints (shared functions, party + global scope)
11. Dashboard endpoints
12. Frontend in same order
