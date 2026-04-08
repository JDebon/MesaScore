# MesaScore — Project Context

## What is this

MesaScore is a self-hosted, multi-tenant web app to track board game sessions. Users register publicly (with email confirmation), create **Parties** (friend groups), and invite others to join. Each party tracks its own sessions and stats. A user can belong to multiple parties. There is no global admin — admin status is per-party (the creator of a party is its admin).

## Methodology

This project is built using **Spec-Driven Development (SDD)**. Specs are written first, implementation follows the specs. Do not implement features that are not specced, and do not deviate from the spec without updating it first.

All specs live in `docs/`:
- `docs/01-data-model/spec.md` + `technical.md` — entities, relationships, PostgreSQL schema
- `docs/02-features/spec.md` + `technical.md` — user stories, acceptance criteria, frontend/backend tasks
- `docs/03-api/spec.md` + `technical.md` — REST API contract, Go implementation tasks

When starting a new task, read the relevant spec first.

## Tech Stack

| Layer | Choice |
|---|---|
| Frontend | SvelteKit (Node adapter, mobile-first) |
| Backend | Go 1.23+ with Chi router |
| Database | PostgreSQL 16 |
| Auth | JWT (golang-jwt/jwt v5) + bcrypt |
| Email | gomail.v2 via SMTP |
| Reverse proxy | Caddy 2 |
| Deployment | Docker Compose on homelab |
| BGG integration | BGG XML API 2 (no auth required) |

## Roles

- **User**: any registered and email-verified user. Can create parties, manage their own game collection, view data for parties they belong to.
- **Party Admin**: the creator of a party (or whoever ownership was transferred to). Can log sessions, manage members, and invite users in their party. There is exactly one admin per party.

There is no global admin role. Ownership can be transferred. Admin cannot leave a party without transferring first.

## Data Model (summary)

Eight tables: `users`, `parties`, `party_members`, `party_invites`, `games`, `user_games`, `sessions`, `session_participants`.

Key design decisions:
- **Game library is global** — any user can add games to the shared catalog. Each user has their own collection (`user_games`) and can use their games in any party they belong to.
- **Sessions are party-scoped** — `sessions.party_id` ties every session to a party.
- **Admin is stored on the party** — `parties.admin_user_id`, not a role field on the member.
- **Invite system** — two mechanisms: invite link (via `parties.invite_code`) and direct username-search invite (`party_invites` table).
- Session types: `competitive`, `team`, `cooperative`, `score`.

See `docs/01-data-model/spec.md` for full entity definitions and design decisions.

## Stats

Stats exist in two scopes:
- **Party scope**: filtered to one party's sessions. Accessed by party members.
- **Global scope**: all sessions a user participated in across all parties. Shown on user profile.

Both scopes use the same underlying query functions — party scope passes an optional `partyID` filter.

## Features (summary)

- **Auth**: public registration with email confirmation, login, profile management.
- **Parties**: create, invite (link or username), accept/decline, remove members, leave, transfer ownership, regenerate invite link.
- **Game catalog**: shared global catalog, any user can add games via BGG search or manually.
- **Collections**: each user manages their own game collection globally; used across any party.
- **Session logging**: party admin logs sessions with flexible result types (competitive/team/co-op/score).
- **Stats**: party leaderboard, per-game stats, player global stats, player party stats, win streaks, nemesis, punching bag, head-to-head.
- **Dashboard**: global user dashboard (all parties + pending invites) and per-party dashboard.

**Explicitly excluded from v1:** achievements, push/email notifications for sessions, password reset, session media, public party discovery, multiple party admins, BGG collection import, real-time updates, party deletion.

## Environment Variables

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

For local development: use `mailpit` (Docker) as SMTP catcher — add to Docker Compose with `profiles: ["dev"]`.

## Docker Compose Services

`postgres`, `backend`, `frontend`, `caddy`, (`mailpit` for dev only)

Deploy: `docker compose up -d --build`

## BGG API Notes

- Base URL: `https://boardgamegeek.com/xmlapi2`
- Search: `GET /search?query=<q>&type=boardgame`
- Game detail: `GET /thing?id=<bgg_id>&stats=1`
- No API key required.
- The `/thing` endpoint may return HTTP 202 on first call — retry once after 2 seconds.
- BGG data is cached in the `games` table (`bgg_fetched_at`). Any user who owns the game can trigger a refresh.

## Frontend Route Structure

```
/                          → global dashboard (parties list + pending invites)
/register                  → registration form
/login                     → login form
/verify-email              → email verification handler
/join/:code                → party join preview (works pre-login)
/games                     → user's own game collection
/games/:id                 → game detail
/games/new                 → add game flow
/users/:id                 → user profile + global stats
/parties/new               → create party
/parties/:id               → party dashboard
/parties/:id/sessions      → session history
/parties/:id/sessions/new  → log session (admin only)
/parties/:id/leaderboard   → party leaderboard
/parties/:id/members       → members + invites
/parties/:id/settings      → party settings (admin only)
/parties/:id/users/:uid    → player stats in party context
```

## Implementation Order

1. DB migrations
2. Register + email verification + login
3. JWT + party middleware
4. User profile + collection endpoints
5. Party CRUD + join via invite link
6. Invite system (username search + accept/decline)
7. Member management (remove, leave, transfer)
8. Game catalog + BGG proxy
9. Session CRUD + validation (party-scoped)
10. Stats (party + global scope)
11. Dashboard endpoints
12. Frontend (same order)
