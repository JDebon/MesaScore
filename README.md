# MesaScore

A self-hosted, multi-tenant web app for tracking board game sessions with friends.

## What is MesaScore?

MesaScore lets you create **Parties** — groups of friends — and log your board game sessions together. Each party has its own session history, leaderboard, and stats. A user can belong to multiple parties.

Key highlights:

- **Party-scoped tracking** — sessions, stats, and leaderboards are isolated per party.
- **Global game catalog** — any user can add games; everyone shares the catalog.
- **Personal collections** — each user maintains their own game collection, usable across all their parties.
- **Flexible session types** — competitive, team, cooperative, and score-based sessions.
- **Invite system** — join parties via invite link or direct username search.
- **Stats** — win streaks, nemesis, punching bag, head-to-head records, and more.
- **BGG integration** — search and import games directly from BoardGameGeek.

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | SvelteKit (Node adapter, mobile-first) |
| Backend | Go 1.23+ with Chi router |
| Database | PostgreSQL 16 |
| Auth | JWT + bcrypt |
| Email | gomail.v2 via SMTP |
| Reverse proxy | Caddy 2 |
| Deployment | Docker Compose |

## Getting Started

### Prerequisites

- Docker and Docker Compose

### Setup

1. Clone the repository:

   ```bash
   git clone <repo-url>
   cd MesaScore
   ```

2. Create a `.env` file from the example:

   ```bash
   cp .env.example .env
   ```

3. Edit `.env` with your configuration (see [Environment Variables](#environment-variables)).

4. Start the stack:

   ```bash
   docker compose up -d --build
   ```

   For local development with email catching (Mailpit):

   ```bash
   docker compose --profile dev up -d --build
   ```

The app will be available at `http://localhost`. Mailpit (email catcher) runs at `http://localhost:8025` in dev mode.

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `POSTGRES_PASSWORD` | PostgreSQL password | `changeme` |
| `JWT_SECRET` | Secret key for signing JWTs | `dev-secret` |
| `BASE_URL` | Public base URL of the app | `http://localhost` |
| `ALLOWED_ORIGIN` | CORS allowed origin | `http://localhost:5173` |
| `SMTP_HOST` | SMTP server hostname | `mailpit` |
| `SMTP_PORT` | SMTP server port | `1025` |
| `SMTP_USER` | SMTP username | *(empty)* |
| `SMTP_PASS` | SMTP password | *(empty)* |
| `SMTP_FROM` | From address for emails | `MesaScore <noreply@mesascore.local>` |
| `BGG_API_TOKEN` | BoardGameGeek API token | *(empty)* |

> **Important:** Never commit your `.env` file. Change `JWT_SECRET` and `POSTGRES_PASSWORD` from their defaults in any non-local deployment.

## How It Was Built — Spec-Driven Development with Claude Code

MesaScore was built using **Spec-Driven Development (SDD)** — every feature was specced before any code was written. The implementation was done entirely with **[Claude Code](https://claude.ai/code)**, Anthropic's AI coding assistant.

The workflow:

1. **Write the spec** — features, data model, and API contract are defined as Markdown documents in `docs/`.
2. **Review and approve** — specs are reviewed before implementation starts.
3. **Implement against the spec** — Claude Code reads the relevant spec and implements it. No feature is built without a corresponding spec.
4. **Specs are the source of truth** — if the code and the spec disagree, the spec wins (or the spec is updated deliberately).

This approach keeps the codebase intentional and auditable. The `docs/` directory contains the full specification:

```
docs/
├── 01-data-model/    # Entities, relationships, PostgreSQL schema
├── 02-features/      # User stories and acceptance criteria
├── 03-api/           # REST API contract
└── 04-frontend/      # Frontend routes and component specs
```

## Project Structure

```
MesaScore/
├── backend/          # Go API server
│   ├── cmd/          # Entry point
│   ├── internal/     # App logic (handlers, models, auth, stats…)
│   └── migrations/   # SQL migrations
├── frontend/         # SvelteKit app
│   └── src/
│       ├── lib/      # Shared components and utilities
│       └── routes/   # Page components
├── docs/             # Specs (source of truth)
├── docker-compose.yml
└── Caddyfile
```

## License

MIT
