# API Specification

## Overview

MesaScore exposes a JSON REST API consumed exclusively by its own SvelteKit frontend. All endpoints are prefixed with `/api`.

**Base URL (production):** `https://<homelab-domain>/api`
**Base URL (development):** `http://localhost:8080`

---

## Authentication

All endpoints except the `/api/auth/*` group and `GET /api/parties/join/:invite_code` require a valid JWT.

```
Authorization: Bearer <token>
```

JWT payload:
```json
{
  "sub": "<user_uuid>",
  "exp": <unix_timestamp>
}
```

**Standard error responses:**
- `401 Unauthorized` — missing or invalid token.
- `403 Forbidden` — valid token but insufficient permissions.

Token refresh: if the token expires within 24 hours, a refreshed token is returned in the `X-New-Token` response header.

**Party authorization:**
- "Party member" — user must have a row in `party_members` for this party.
- "Party admin" — user must be the `admin_user_id` of this party.

---

## Common Conventions

- All IDs are UUID v4.
- Timestamps: ISO 8601, e.g. `2025-10-15T20:30:00Z`.
- Paginated list responses:
  ```json
  { "data": [...], "total": 42, "page": 1, "per_page": 20 }
  ```
- Error responses:
  ```json
  { "error": "human-readable message" }
  ```
- Validation errors:
  ```json
  { "error": "validation failed", "fields": { "field_name": "message" } }
  ```

---

## Endpoints

---

### Auth

#### `POST /api/auth/register`
No auth required.

Request:
```json
{
  "username": "string",
  "display_name": "string",
  "email": "string",
  "password": "string"
}
```

Response `201`:
```json
{ "message": "Check your email to verify your account" }
```

Errors: `409` if email or username taken. `400` on validation failure.

---

#### `GET /api/auth/verify-email?token=<token>`
No auth required.

Response `200`:
```json
{ "message": "Email verified successfully" }
```

Errors: `400` if token invalid or expired.

---

#### `POST /api/auth/resend-verification`
No auth required.

Request:
```json
{ "email": "string" }
```

Response `200`: always succeeds (does not reveal whether email exists).

---

#### `GET /api/auth/check-username?username=<val>`
No auth required. Used for real-time availability check on registration.

Response `200`:
```json
{ "available": true | false }
```

---

#### `POST /api/auth/login`
No auth required.

Request:
```json
{ "email": "string", "password": "string" }
```

Response `200`:
```json
{
  "token": "string",
  "user": {
    "id": "uuid",
    "username": "string",
    "display_name": "string",
    "avatar_url": "string | null"
  }
}
```

Errors: `401` on bad credentials. `403` with `{ "error": "email_not_verified" }` if not verified.

---

### Users

#### `GET /api/users/me`
Auth: any user.

Response `200`:
```json
{
  "id": "uuid",
  "username": "string",
  "display_name": "string",
  "avatar_url": "string | null",
  "created_at": "timestamp"
}
```

#### `PATCH /api/users/me`
Auth: any user.

Request (all optional):
```json
{
  "display_name": "string",
  "avatar_url": "string | null"
}
```

Response `200`: updated user object.

---

#### `GET /api/users/:id`
Auth: any user.

Response `200`:
```json
{
  "id": "uuid",
  "username": "string",
  "display_name": "string",
  "avatar_url": "string | null",
  "created_at": "timestamp"
}
```

Errors: `404`.

---

#### `GET /api/users/search?q=<query>`
Auth: any user. Used for invite-by-username search.

Response `200`:
```json
[
  { "id": "uuid", "username": "string", "display_name": "string", "avatar_url": "string | null" }
]
```

Returns up to 10 results. Does not return the requesting user.

---

#### `GET /api/users/me/dashboard`
Auth: any user.

Response `200`:
```json
{
  "parties": [
    {
      "id": "uuid",
      "name": "string",
      "member_count": "integer",
      "last_session_at": "timestamp | null"
    }
  ],
  "pending_invites": [
    {
      "id": "uuid",
      "party": { "id": "uuid", "name": "string" },
      "invited_by": { "id": "uuid", "display_name": "string" },
      "created_at": "timestamp"
    }
  ],
  "global_stats": {
    "total_sessions": "integer",
    "total_wins": "integer",
    "current_streak": "integer"
  }
}
```

---

#### `GET /api/users/me/invites`
Auth: any user.

Response `200`: same as `pending_invites` array from dashboard.

---

#### `GET /api/users/:id/stats`
Auth: any user. Global stats (all parties).

Response `200`:
```json
{
  "user": { "id": "uuid", "display_name": "string", "avatar_url": "string | null" },
  "total_sessions": "integer",
  "total_wins": "integer",
  "win_rate": "number",
  "current_streak": "integer",
  "best_streak": "integer",
  "most_played_game": { "id": "uuid", "name": "string", "session_count": "integer" } | null,
  "best_win_rate_game": { "id": "uuid", "name": "string", "win_rate": "number" } | null,
  "nemesis": { "id": "uuid", "display_name": "string", "losses_against": "integer" } | null,
  "punching_bag": { "id": "uuid", "display_name": "string", "wins_against": "integer" } | null,
  "per_game": [
    {
      "game": { "id": "uuid", "name": "string", "cover_image_url": "string | null" },
      "sessions": "integer",
      "wins": "integer",
      "win_rate": "number"
    }
  ],
  "head_to_head": [
    {
      "opponent": { "id": "uuid", "display_name": "string", "avatar_url": "string | null" },
      "sessions_together": "integer",
      "this_user_wins": "integer",
      "opponent_wins": "integer"
    }
  ]
}
```

---

#### `GET /api/users/:id/collection`
Auth: any user.

Response `200`:
```json
[
  {
    "game_id": "uuid",
    "name": "string",
    "cover_image_url": "string | null",
    "added_at": "timestamp"
  }
]
```

---

#### `POST /api/users/me/collection`
Auth: any user.

Request:
```json
{ "game_id": "uuid" }
```

Response `201`: no body.

Errors: `409` if already in collection. `404` if game not found.

---

#### `DELETE /api/users/me/collection/:game_id`
Auth: any user.

Response `204`: no body.

Errors: `404` if not in collection.

---

### Parties

#### `POST /api/parties`
Auth: any user.

Request:
```json
{
  "name": "string",
  "description": "string | null"
}
```

Response `201`:
```json
{ "id": "uuid", "invite_code": "string" }
```

---

#### `GET /api/parties/:id`
Auth: party member.

Response `200`:
```json
{
  "id": "uuid",
  "name": "string",
  "description": "string | null",
  "admin": { "id": "uuid", "display_name": "string", "avatar_url": "string | null" },
  "invite_code": "string",
  "member_count": "integer",
  "created_at": "timestamp"
}
```

Errors: `403` if not a member. `404` if party not found.

---

#### `PATCH /api/parties/:id`
Auth: party admin.

Request (all optional):
```json
{
  "name": "string",
  "description": "string | null"
}
```

Response `200`: updated party object.

---

#### `POST /api/parties/:id/regenerate-invite`
Auth: party admin.

Response `200`:
```json
{ "invite_code": "string" }
```

---

#### `GET /api/parties/join/:invite_code`
No auth required. Preview before joining.

Response `200`:
```json
{
  "party": { "id": "uuid", "name": "string", "member_count": "integer" }
}
```

Errors: `404` if invite code not found.

---

#### `POST /api/parties/join/:invite_code`
Auth: any user.

Response `200`:
```json
{ "party_id": "uuid" }
```

Errors: `409` if already a member. `404` if invite code not found.

---

#### `GET /api/parties/:id/members`
Auth: party member.

Response `200`:
```json
{
  "members": [
    {
      "id": "uuid",
      "username": "string",
      "display_name": "string",
      "avatar_url": "string | null",
      "is_admin": "boolean",
      "joined_at": "timestamp"
    }
  ],
  "invites": [
    {
      "id": "uuid",
      "invited_user": { "id": "uuid", "username": "string", "display_name": "string" },
      "status": "pending | accepted | declined",
      "created_at": "timestamp"
    }
  ]
}
```

Note: `invites` is only populated for the party admin.

---

#### `DELETE /api/parties/:id/members/:user_id`
Auth: party admin.

Response `204`: no body.

Errors: `400` if trying to remove self (admin). `404` if user not a member.

---

#### `POST /api/parties/:id/leave`
Auth: party member (non-admin).

Response `204`: no body.

Errors: `400` if user is the admin (must transfer first).

---

#### `POST /api/parties/:id/transfer-admin`
Auth: party admin.

Request:
```json
{ "new_admin_user_id": "uuid" }
```

Response `200`:
```json
{ "message": "Ownership transferred" }
```

Errors: `400` if target is not a party member.

---

#### `POST /api/parties/:id/invites`
Auth: party admin.

Request:
```json
{ "user_id": "uuid" }
```

Response `201`: no body.

Errors: `409` if already a member or pending invite exists. `404` if user not found.

---

#### `POST /api/parties/:id/invites/:invite_id/accept`
Auth: the invited user only.

Response `200`:
```json
{ "party_id": "uuid" }
```

Errors: `403` if not the invited user. `404` if invite not found.

---

#### `POST /api/parties/:id/invites/:invite_id/decline`
Auth: the invited user only.

Response `204`: no body.

Errors: `403` if not the invited user.

---

#### `GET /api/parties/:id/dashboard`
Auth: party member.

Response `200`:
```json
{
  "total_sessions": "integer",
  "total_unique_games": "integer",
  "total_members": "integer",
  "current_leader": {
    "user": { "id": "uuid", "display_name": "string", "avatar_url": "string | null" },
    "wins": "integer"
  } | null,
  "most_played_game": {
    "id": "uuid", "name": "string", "cover_image_url": "string | null", "session_count": "integer"
  } | null,
  "sessions_per_month": [
    { "month": "2025-10", "count": "integer" }
  ],
  "recent_sessions": [
    {
      "id": "uuid",
      "game": { "id": "uuid", "name": "string", "cover_image_url": "string | null" },
      "played_at": "timestamp",
      "session_type": "string",
      "winners": [{ "id": "uuid", "display_name": "string" }]
    }
  ]
}
```

---

### Games

#### `GET /api/games`
Auth: any user.

Query params: `q` (name search), `sort` (`name` | `rating` | `sessions`).

Response `200`:
```json
[
  {
    "id": "uuid",
    "bgg_id": "integer | null",
    "name": "string",
    "cover_image_url": "string | null",
    "min_players": "integer | null",
    "max_players": "integer | null",
    "bgg_rating": "number | null",
    "session_count": "integer",
    "in_my_collection": "boolean"
  }
]
```

---

#### `GET /api/games/:id`
Auth: any user.

Response `200`:
```json
{
  "id": "uuid",
  "bgg_id": "integer | null",
  "name": "string",
  "description": "string | null",
  "cover_image_url": "string | null",
  "min_players": "integer | null",
  "max_players": "integer | null",
  "bgg_rating": "number | null",
  "bgg_fetched_at": "timestamp | null",
  "added_by": { "id": "uuid", "display_name": "string" },
  "created_at": "timestamp",
  "session_count": "integer",
  "in_my_collection": "boolean",
  "owners": [
    { "id": "uuid", "display_name": "string", "avatar_url": "string | null" }
  ]
}
```

Errors: `404`.

---

#### `POST /api/games`
Auth: any user.

Request:
```json
{
  "bgg_id": "integer | null",
  "name": "string",
  "description": "string | null",
  "cover_image_url": "string | null",
  "min_players": "integer | null",
  "max_players": "integer | null"
}
```

Response `201`:
```json
{ "id": "uuid" }
```

Errors: `409` if `bgg_id` already in catalog.

---

#### `PATCH /api/games/:id`
Auth: any user who owns the game or added it.

Request (all optional):
```json
{
  "name": "string",
  "description": "string | null",
  "cover_image_url": "string | null",
  "min_players": "integer | null",
  "max_players": "integer | null"
}
```

Response `200`: updated game object.

---

#### `POST /api/games/:id/bgg-refresh`
Auth: any user who owns the game.

Response `200`: updated game object.

Errors: `400` if game has no `bgg_id`. `502` if BGG unreachable.

---

#### `GET /api/bgg/search?q=<query>`
Auth: any user.

Response `200`:
```json
[
  {
    "bgg_id": "integer",
    "name": "string",
    "year": "integer | null",
    "thumbnail_url": "string | null"
  }
]
```

Errors: `502` if BGG unreachable.

---

#### `GET /api/parties/:id/available-games`
Auth: party member. Returns union of all member collections — games available to use in a session.

Response `200`:
```json
[
  {
    "id": "uuid",
    "name": "string",
    "cover_image_url": "string | null",
    "owners": [
      { "id": "uuid", "display_name": "string" }
    ]
  }
]
```

---

### Sessions

#### `GET /api/parties/:id/sessions`
Auth: party member. Paginated.

Query params: `game_id`, `user_id`, `type`, `from`, `to`, `page`, `per_page`.

Response `200` (paginated):
```json
{
  "data": [
    {
      "id": "uuid",
      "game": { "id": "uuid", "name": "string", "cover_image_url": "string | null" },
      "session_type": "string",
      "played_at": "timestamp",
      "duration_minutes": "integer | null",
      "winners": [{ "id": "uuid", "display_name": "string", "avatar_url": "string | null" }],
      "participant_count": "integer"
    }
  ],
  "total": "integer",
  "page": "integer",
  "per_page": "integer"
}
```

---

#### `GET /api/parties/:id/sessions/:session_id`
Auth: party member.

Response `200`:
```json
{
  "id": "uuid",
  "game": { "id": "uuid", "name": "string", "cover_image_url": "string | null" },
  "session_type": "competitive | team | cooperative | score",
  "played_at": "timestamp",
  "duration_minutes": "integer | null",
  "notes": "string | null",
  "brought_by": { "id": "uuid", "display_name": "string" } | null,
  "created_by": { "id": "uuid", "display_name": "string" },
  "created_at": "timestamp",
  "participants": [
    {
      "user": { "id": "uuid", "display_name": "string", "avatar_url": "string | null" },
      "team_name": "string | null",
      "rank": "integer | null",
      "score": "number | null",
      "result": "win | loss | draw | null"
    }
  ]
}
```

Errors: `404`. `403` if not a party member.

---

#### `POST /api/parties/:id/sessions`
Auth: party member.

Request:
```json
{
  "game_id": "uuid",
  "session_type": "competitive | team | cooperative | score",
  "played_at": "timestamp",
  "duration_minutes": "integer | null",
  "notes": "string | null",
  "brought_by_user_id": "uuid | null",
  "participants": [
    {
      "user_id": "uuid",
      "team_name": "string | null",
      "rank": "integer | null",
      "score": "number | null",
      "result": "win | loss | draw | null"
    }
  ]
}
```

Response `201`:
```json
{ "id": "uuid" }
```

Errors: `400` with field errors. `403` if not party admin. `404` if game or user not found.

---

#### `PATCH /api/parties/:id/sessions/:session_id`
Auth: party admin.

Request: same shape as POST, all fields optional. Providing `participants` replaces the full list.

Response `200`: updated session detail object.

---

#### `DELETE /api/parties/:id/sessions/:session_id`
Auth: party admin.

Response `204`: no body.

Errors: `404`.

---

### Stats

#### `GET /api/parties/:id/stats/leaderboard`
Auth: party member.

Query params: `sort` (`wins` | `win_rate` | `sessions`, default `wins`).

Response `200`:
```json
[
  {
    "user": { "id": "uuid", "display_name": "string", "avatar_url": "string | null" },
    "wins": "integer",
    "sessions": "integer",
    "win_rate": "number"
  }
]
```

---

#### `GET /api/parties/:id/stats/games/:game_id`
Auth: party member.

Response `200`:
```json
{
  "game": { "id": "uuid", "name": "string" },
  "total_sessions": "integer",
  "last_played_at": "timestamp | null",
  "sessions_per_month": [{ "month": "2025-10", "count": "integer" }],
  "leaderboard": [
    {
      "user": { "id": "uuid", "display_name": "string", "avatar_url": "string | null" },
      "wins": "integer",
      "sessions": "integer",
      "win_rate": "number"
    }
  ]
}
```

---

#### `GET /api/parties/:id/stats/activity`
Auth: party member.

Response `200`:
```json
[{ "month": "2025-10", "count": "integer" }]
```

---

#### `GET /api/parties/:id/users/:user_id/stats`
Auth: party member. Player stats scoped to this party.

Response `200`: same shape as `GET /api/users/:id/stats`, but all metrics filtered to this party's sessions.

---
