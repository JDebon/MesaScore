# Features — Technical Tasks

## Stack Reference
- **Backend**: Go 1.23+ with Chi router
- **Frontend**: SvelteKit (Node adapter, mobile-first)
- **Database**: PostgreSQL 16
- **Auth**: JWT (golang-jwt/jwt v5) + bcrypt
- **Email**: SMTP via `gopkg.in/gomail.v2` (configured with env vars)
- **BGG**: BGG XML API 2
- **Containerization**: Docker Compose

---

## F1 — Authentication

### T2.1 — Email service (backend)
- Package `internal/email` with function `SendVerificationEmail(to, token string) error`.
- Builds verification URL: `<BASE_URL>/verify-email?token=<token>`.
- Sends via SMTP using gomail. Config from env: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM`.
- Email subject: "Confirm your MesaScore account".
- HTML body with a prominent CTA button and plain-text fallback.

### T2.2 — Registration endpoint
- `POST /api/auth/register` — no auth required.
- Accepts `{ username, display_name, email, password }`.
- Validates uniqueness of username and email.
- Hashes password (bcrypt, cost 12).
- Generates `verification_token` (UUID v4), sets `verification_sent_at = now`.
- Inserts user with `email_verified = false`.
- Calls email service to send verification email.
- Returns `201` with `{ message: "Check your email to verify your account" }`.

### T2.3 — Email verification endpoint
- `GET /api/auth/verify-email?token=<token>` — no auth required.
- Finds user by `verification_token`.
- Checks `verification_sent_at` is within 24 hours.
- If valid: sets `email_verified = true`, clears `verification_token` and `verification_sent_at`.
- Returns `200` with `{ message: "Email verified" }`.
- Frontend redirects to login on success.
- Returns `400` if token is invalid or expired.

### T2.4 — Resend verification email endpoint
- `POST /api/auth/resend-verification` — no auth required.
- Accepts `{ email }`.
- If user exists and is unverified: generates new token, updates `verification_sent_at`, sends email.
- Always returns `200` (do not reveal whether email exists).

### T2.5 — Login endpoint
- `POST /api/auth/login` — no auth required.
- Accepts `{ email, password }`.
- Rejects with `403` and `{ error: "email_not_verified" }` if `email_verified = false` (lets frontend show "resend" option).
- Issues JWT with payload `{ sub: user_id, exp }` on success.
- Returns `401` on bad credentials with generic message.

### T2.6 — JWT middleware
- `AuthMiddleware`: validates JWT, attaches `user_id` to request context.
- `PartyAdminMiddleware`: checks that the user is the admin of the party in the URL param.
- `PartyMemberMiddleware`: checks that the user is a member of the party in the URL param.
- Token refresh: if token expires within 24 hours, return new token in `X-New-Token` header.

### T2.7 — Profile update endpoint
- `PATCH /api/users/me` — authenticated.
- Accepts `{ display_name, avatar_url }`.
- Returns updated user object.

### T2.8 — Frontend: auth pages
- `/register` — registration form with real-time username availability check (debounced `GET /api/auth/check-username?username=<val>`).
- `/verify-email` — calls verify endpoint on mount, shows success or error.
- `/login` — standard form. On `email_not_verified` error, shows "Resend verification email" link.
- Auth store: Svelte store holding `{ id, display_name, avatar_url }` decoded from JWT.

---

## F2 — Parties

### T3.1 — Party CRUD endpoints
- `POST /api/parties` — create party. Auth: any verified user. Generates invite code. Creates PartyMember row for creator.
- `GET /api/parties/:id` — party detail. Auth: party member only.
- `PATCH /api/parties/:id` — update name/description. Auth: party admin.
- `GET /api/parties/:id/members` — list members + pending invites. Auth: party member.

### T3.2 — Join via invite link
- `GET /api/parties/join/:invite_code` — no auth required. Returns party name and member count (preview).
- `POST /api/parties/join/:invite_code` — authenticated. Adds user to party (creates PartyMember row). Returns `409` if already a member.

### T3.3 — Invite by username
- `GET /api/users/search?q=<query>` — authenticated. Returns users matching query, excludes the searching user. Frontend filters out existing members and pending invitees.
- `POST /api/parties/:id/invites` — party admin. Body: `{ user_id }`. Creates PartyInvite. Returns `409` if already member or pending invite exists.
- `GET /api/parties/:id/invites` — party admin. Returns all invites with status.

### T3.4 — Invite response endpoints
- `GET /api/users/me/invites` — authenticated. Returns pending invites for the current user.
- `POST /api/parties/:id/invites/:invite_id/accept` — authenticated (invited user only). Creates PartyMember, updates invite status.
- `POST /api/parties/:id/invites/:invite_id/decline` — authenticated (invited user only). Updates invite status.

### T3.5 — Member management endpoints
- `DELETE /api/parties/:id/members/:user_id` — party admin. Removes member (cannot remove self as admin).
- `POST /api/parties/:id/leave` — authenticated party member (non-admin). Removes self from party.
- `POST /api/parties/:id/transfer-admin` — party admin. Body: `{ new_admin_user_id }`. Updates `parties.admin_user_id`. Target must be current member.
- `POST /api/parties/:id/regenerate-invite` — party admin. Generates new invite code. Returns new code.

### T3.6 — Frontend: party pages
- `/` (global dashboard) — list user's parties + pending invites + global quick stats.
- `/parties/new` — create party form.
- `/join/:invite_code` — join preview page (works before login; saves invite code to session, redirects through auth, then auto-joins).
- `/parties/:id` — party dashboard (F6.2).
- `/parties/:id/members` — members list with invite management (admin sees full controls).
- `/parties/:id/settings` — edit party name/description, regenerate invite, transfer ownership, danger zone.

### T3.7 — Frontend: invite pending banner
- If user has pending invites, show a notification badge on the global dashboard.
- Each invite card shows party name, invited by whom, Accept/Decline buttons.

---

## F3 — Game Catalog & Collections

### T4.1 — Game catalog endpoints
- `GET /api/games` — list catalog. Auth: any user. Query params: `q`, `sort`. Returns `in_collection` boolean per game for the requesting user.
- `GET /api/games/:id` — game detail. Auth: any user. Returns owners list (all users globally who own it).
- `POST /api/games` — add game. Auth: any user. Auto-adds to requester's collection.
- `PATCH /api/games/:id` — edit game. Auth: any user who owns it (or added it).
- `POST /api/games/:id/bgg-refresh` — re-fetch BGG data. Auth: any user who owns it.

### T4.2 — BGG proxy endpoint
- `GET /api/bgg/search?q=<query>` — authenticated. No admin restriction.

### T4.3 — Collection endpoints
- `GET /api/users/:id/collection` — list user's collection. Auth: any user.
- `POST /api/users/me/collection` — add game to own collection. Body: `{ game_id }`.
- `DELETE /api/users/me/collection/:game_id` — remove game from own collection.

### T4.4 — Frontend: games pages
- `/games` — user's own collection. Loads `GET /api/users/:id/collection`. Client-side name filter. No sort controls (collection items lack rating/session data).
- `/games/:id` — game detail. Shows "Add to collection" / "Remove from collection" toggle. Shows all global owners.
- `/games/new` — add game flow (BGG search → pre-fill form → submit).

### T4.5 — Frontend: collection on profile
- `/users/:id` — profile page. Includes collection tab.
- Own profile shows "Remove" buttons. Others' profiles are view-only.

---

## F4 — Session Logging

### T5.1 — Session endpoints
- `GET /api/parties/:id/sessions` — list sessions for party. Auth: party member. Supports filters.
- `GET /api/parties/:id/sessions/:session_id` — session detail. Auth: party member.
- `POST /api/parties/:id/sessions` — create session. Auth: party member.
- `PATCH /api/parties/:id/sessions/:session_id` — edit session. Auth: party admin.
- `DELETE /api/parties/:id/sessions/:session_id` — delete session. Auth: party admin.

### T5.2 — Game availability for session
- `GET /api/parties/:id/available-games` — authenticated party member. Returns games that at least one party member owns (union of all member collections). Used in session creation form to populate game dropdown.

### T5.3 — Session validation (same as original, with party scope)
- All participants must be current party members.
- `brought_by_user_id` (if set) must be a party member who owns the game.
- All other validation rules from original spec apply.

### T5.4 — Frontend: log session form
- Located at `/parties/:id/sessions/new`.
- Multi-step form identical to original spec, adapted to use party-scoped game list and member list.

### T5.5 — Frontend: session list and detail pages
- `/parties/:id/sessions` — filterable session list.
- `/parties/:id/sessions/:session_id` — detail view with admin controls.

---

## F5 — Stats & Leaderboards

### T6.1 — Stats endpoints
- `GET /api/parties/:id/stats/leaderboard` — party leaderboard. Auth: party member.
- `GET /api/parties/:id/stats/games/:game_id` — per-game stats within party. Auth: party member.
- `GET /api/parties/:id/stats/activity` — sessions per month for party. Auth: party member.
- `GET /api/users/:id/stats` — global user stats (all parties). Auth: any user.
- `GET /api/parties/:id/users/:user_id/stats` — user stats scoped to one party. Auth: party member.

### T6.2 — Stats query implementation
All stat functions take an optional `partyID` parameter. When provided, queries filter to `sessions.party_id = partyID`. When null, all sessions the user participated in are included.

Implement in `internal/stats/`:
- `Leaderboard(ctx, db, partyID)` — wins/sessions/win_rate per member.
- `GameStats(ctx, db, partyID, gameID)` — per-game leaderboard.
- `PlayerStats(ctx, db, userID, partyID *uuid.UUID)` — full player stats, optionally scoped.
- `ActivityPerMonth(ctx, db, partyID)` — sessions per month.
- `WinStreak(ctx, db, userID, partyID *uuid.UUID)` — current and best streak.
- `Nemesis(ctx, db, userID, partyID *uuid.UUID)` — most losses against.
- `PunchingBag(ctx, db, userID, partyID *uuid.UUID)` — most wins against.
- `HeadToHead(ctx, db, userID, opponentID uuid.UUID, partyID *uuid.UUID)` — shared sessions.

### T6.3 — Frontend: stats pages
- `/parties/:id/leaderboard` — party leaderboard.
- `/users/:id` (stats tab) — global player stats.
- `/parties/:id/users/:user_id` — player stats in party context (shows both tabs: party and global).

---

## F6 — Dashboard

### T7.1 — Dashboard endpoints
- `GET /api/users/me/dashboard` — global user dashboard data. Returns: parties list, pending invites, global quick stats.
- `GET /api/parties/:id/dashboard` — party dashboard. Returns: recent sessions, party totals, current leader, most played game, sessions per month.

### T7.2 — Frontend: global dashboard (`/`)
- Shows party cards (tap to enter party).
- Pending invites section with Accept/Decline.
- Global stats strip (total sessions, total wins, current streak).
- "Create party" FAB (floating action button, mobile-friendly).

### T7.3 — Frontend: party dashboard (`/parties/:id`)
- Party name header with settings gear (admin only).
- Stats strip (sessions, games, members).
- Current leader highlight.
- Recent sessions feed (last 5, tappable).
- Sessions per month bar chart.
- Bottom nav (mobile): Dashboard / Sessions / Games / Members / Leaderboard.

---

## Infrastructure

### T8.1 — Backend project structure

```
backend/
  cmd/server/main.go
  internal/
    auth/           # JWT, bcrypt
    bgg/            # BGG XML client
    db/             # pgx pool, migration runner
    email/          # SMTP email sender
    handlers/       # HTTP handlers (auth, users, parties, games, sessions, stats)
    middleware/     # AuthMiddleware, PartyMemberMiddleware, PartyAdminMiddleware
    models/         # Go structs
    stats/          # stat query functions
    validator/      # session validation
  migrations/
  scripts/seed.go
  Dockerfile
```

### T8.2 — Frontend project structure

```
frontend/
  src/
    lib/
      api/          # typed fetch wrappers per resource
      components/   # Card, Table, Avatar, Chart, Modal, BottomNav
      stores/       # auth store, current party store
    routes/
      (auth)/       # unauthenticated layout
        login/
        register/
        verify-email/
      join/[code]/  # invite link join page (outside auth layout)
      (app)/        # authenticated layout
        +layout.svelte
        (global)/   # no party context
          +page.svelte          # global dashboard
          games/                # catalog
          users/[id]/           # user profile + global stats
        parties/
          new/
          [id]/                 # party layout (bottom nav)
            +page.svelte        # party dashboard
            sessions/
            leaderboard/
            members/
            settings/
            users/[userId]/     # player stats in party context
  Dockerfile
```

### T8.3 — Environment variables (updated)

```
POSTGRES_DB=mesascore
POSTGRES_USER=mesascore
POSTGRES_PASSWORD=changeme
JWT_SECRET=changeme
BASE_URL=http://localhost
ALLOWED_ORIGIN=http://localhost:5173
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@example.com
SMTP_PASS=changeme
SMTP_FROM=MesaScore <noreply@example.com>
```

For homelab without external SMTP, a local mail relay (e.g., `ixdotai/smtp` Docker image) can be added as a fifth Docker Compose service.

### T8.4 — Docker Compose services

`postgres`, `backend`, `frontend`, `caddy`, optionally `smtp` (local relay).

### T8.5 — Implementation order (updated)

1. DB migrations
2. Register + verify email + login endpoints
3. JWT middleware
4. User profile endpoints
5. Party CRUD + membership endpoints
6. Invite system (link + username)
7. Game catalog + BGG proxy
8. Collection endpoints
9. Session CRUD + validation (party-scoped)
10. Stats endpoints (party + global)
11. Dashboard endpoints
12. Frontend (same order)
