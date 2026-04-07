# Data Model Specification

## Overview

MesaScore is a multi-tenant board game tracking app. Users register publicly (with email confirmation), create **Parties** (their friend groups), and invite others to join. Each party tracks its own sessions. A user can belong to multiple parties. Game collections belong to individual users globally — not to a party — and can be used in any party the user belongs to.

---

## Entities

### User

Represents any registered person using the app.

| Field | Type | Constraints | Description |
|---|---|---|---|
| id | UUID | PK | Unique identifier |
| username | string | UNIQUE, NOT NULL, max 50 chars | Public handle |
| display_name | string | NOT NULL, max 100 chars | Name shown in UI |
| email | string | UNIQUE, NOT NULL | Used for login and verification |
| password_hash | string | NOT NULL | bcrypt hash |
| email_verified | boolean | NOT NULL, default false | Whether email has been confirmed |
| verification_token | string | UNIQUE, nullable | Token sent in confirmation email |
| verification_sent_at | timestamp | nullable | When the last verification email was sent |
| avatar_url | string | nullable, max 500 chars | Profile picture URL |
| created_at | timestamp | NOT NULL, default now | Account creation date |

**Rules:**
- A user cannot log in until `email_verified` is true.
- `verification_token` is cleared after successful verification.
- Verification tokens expire 24 hours after `verification_sent_at`.
- A new verification email can be requested if the token is expired, which generates a new token.
- Users are never deleted — sessions and party records reference them permanently.
- There is no global "admin" role. Admin status is per-party (see Party).

---

### Party

Represents a friend group. One user is always the admin.

| Field | Type | Constraints | Description |
|---|---|---|---|
| id | UUID | PK | Unique identifier |
| name | string | NOT NULL, max 100 chars | Group name |
| description | text | nullable | Optional description |
| admin_user_id | UUID | FK → users(id), NOT NULL | Current admin of the party |
| invite_code | string | UNIQUE, NOT NULL, 32 chars | Used to generate join links |
| created_at | timestamp | NOT NULL, default now | When the party was created |

**Rules:**
- Any verified user can create a party. The creator becomes the admin.
- `invite_code` is a randomly generated token. The join URL is `<base_url>/join/<invite_code>`.
- The admin can regenerate the invite code at any time, invalidating all old links.
- The admin can also invite specific users by username search (see PartyInvite).
- The admin is always a member of the party.
- The admin cannot leave the party without first transferring ownership to another member.
- Ownership transfer changes `admin_user_id` to the new admin's ID.
- A party cannot be deleted if it has sessions. (Out of scope for v1 — parties are permanent.)

---

### PartyMember

Tracks which users belong to which parties.

| Field | Type | Constraints | Description |
|---|---|---|---|
| id | UUID | PK | Unique identifier |
| party_id | UUID | FK → parties(id), NOT NULL | The party |
| user_id | UUID | FK → users(id), NOT NULL | The member |
| joined_at | timestamp | NOT NULL, default now | When they joined |

**Unique constraint:** (party_id, user_id).

**Rules:**
- Admin membership is implicit — the party's `admin_user_id` is always a member.
- A user who accepts an invite gets a PartyMember row.
- A user who joins via invite link gets a PartyMember row immediately.
- The admin can remove any non-admin member from the party (removes their PartyMember row).
- A non-admin member can leave the party themselves (removes their own row).
- Removing a member does not delete their historical session participation records.

---

### PartyInvite

Tracks direct (username-based) invitations sent to specific users.

| Field | Type | Constraints | Description |
|---|---|---|---|
| id | UUID | PK | Unique identifier |
| party_id | UUID | FK → parties(id), NOT NULL | The party |
| invited_user_id | UUID | FK → users(id), NOT NULL | User being invited |
| invited_by_user_id | UUID | FK → users(id), NOT NULL | Admin who sent the invite |
| status | enum | NOT NULL, default `pending` | `pending`, `accepted`, `declined` |
| created_at | timestamp | NOT NULL, default now | When the invite was sent |

**Unique constraint:** (party_id, invited_user_id) where status = `pending` — one pending invite per user per party.

**Rules:**
- Only the party admin can send invitations.
- If the user is already a member, the invite is rejected.
- Accepting an invite creates a PartyMember row and sets status to `accepted`.
- Declining sets status to `declined`. The admin can re-invite a declined user.
- Invite links (via `invite_code`) bypass this table — they grant membership directly.
- A user already invited cannot be invited again while a pending invite exists.

---

### Game

A shared global catalog of board games. Any verified user can add a game. Functions as a BGG cache.

| Field | Type | Constraints | Description |
|---|---|---|---|
| id | UUID | PK | Unique identifier |
| bgg_id | integer | UNIQUE, nullable | BoardGameGeek game ID |
| name | string | NOT NULL, max 255 chars | Game name |
| description | text | nullable | Game description |
| cover_image_url | string | nullable, max 500 chars | Cover art |
| min_players | integer | nullable | Minimum player count |
| max_players | integer | nullable | Maximum player count |
| bgg_rating | decimal(4,2) | nullable | BGG community rating |
| bgg_fetched_at | timestamp | nullable | When BGG data was last synced |
| added_by_user_id | UUID | FK → users(id), NOT NULL | Who first added this game |
| created_at | timestamp | NOT NULL, default now | When added to the catalog |

**Rules:**
- Any verified user can add a game to the catalog.
- When a user adds a game (by BGG search or manually), it is automatically added to their personal collection.
- Any user who has the game in their collection can trigger a BGG data refresh.
- A game in the catalog is not associated with any specific party.
- Duplicate BGG IDs are rejected — if the game is already in the catalog, the user is prompted to add it to their collection instead.
- A game cannot be removed from the catalog if sessions reference it.

---

### UserGame

A user's personal game collection. Global — not scoped to a party.

| Field | Type | Constraints | Description |
|---|---|---|---|
| id | UUID | PK | Unique identifier |
| user_id | UUID | FK → users(id), NOT NULL | Owner |
| game_id | UUID | FK → games(id), NOT NULL | Game owned |
| added_at | timestamp | NOT NULL, default now | When added to collection |

**Unique constraint:** (user_id, game_id).

**Rules:**
- Any verified user manages their own collection.
- A user can add any game from the catalog to their collection.
- A user can remove a game from their collection without affecting session history.
- A user can use any game in their collection in sessions of any party they belong to.

---

### Session

A single play session of one game, scoped to a party.

| Field | Type | Constraints | Description |
|---|---|---|---|
| id | UUID | PK | Unique identifier |
| party_id | UUID | FK → parties(id), NOT NULL | Which party this session belongs to |
| game_id | UUID | FK → games(id), NOT NULL | Game played |
| session_type | enum | NOT NULL | `competitive`, `team`, `cooperative`, `score` |
| played_at | timestamp | NOT NULL | When the session took place |
| duration_minutes | integer | nullable | How long the session lasted |
| brought_by_user_id | UUID | FK → users(id), nullable | Whose physical copy was used |
| notes | text | nullable | Free-text notes |
| created_by | UUID | FK → users(id), NOT NULL | Who logged the session (party admin) |
| created_at | timestamp | NOT NULL, default now | When the session was logged |

**Rules:**
- Only the party admin can create, edit, or delete sessions within their party.
- All participants must be members of the party.
- `brought_by_user_id` must be a party member who has the game in their collection, or null.
- A session must have at least 2 participants.
- `played_at` cannot be in the future.

---

### SessionParticipant

Records each user's result within a session.

| Field | Type | Constraints | Description |
|---|---|---|---|
| id | UUID | PK | Unique identifier |
| session_id | UUID | FK → sessions(id) ON DELETE CASCADE | Session |
| user_id | UUID | FK → users(id), NOT NULL | Participant |
| team_name | string | nullable, max 100 chars | Team grouping for team games |
| rank | integer | nullable | Final placement (1 = best) |
| score | decimal(10,2) | nullable | Raw score for score-based games |
| result | enum | nullable | `win`, `loss`, `draw` — for cooperative games |

**Unique constraint:** (session_id, user_id).

**Rules per session type:** same as original spec — see `design decisions` section below.

---

## Relationships

```
User ──< PartyMember >── Party
User ──< PartyInvite (invited)
User ──< PartyInvite (invited_by)
Party ── admin_user_id ──> User
Party ──< Session
User ──< UserGame >── Game
Session ──< SessionParticipant >── User
Session ──> Game
Session ── brought_by_user_id ──> User
```

---

## Design Decisions

### Why is admin status stored on Party rather than PartyMember?

A party always has exactly one admin. Storing `admin_user_id` on `Party` makes this constraint explicit and avoids edge cases (e.g., accidentally having two admins or zero admins). Querying "is this user the admin of this party" is a simple equality check.

### Why a global game catalog instead of per-party libraries?

Users own games, not parties. The same physical game can be used across any party the owner belongs to. A global catalog backed by BGG also avoids duplicate entries for the same game (e.g., everyone adding "Wingspan" separately). The catalog is additive — anyone can contribute, no one can remove.

### Why keep SessionParticipant rules identical to original spec?

The session result model (competitive/team/cooperative/score) is independent of the party architecture. All the original design decisions about session types still apply.

### Party stats vs. global user stats

- **Party stats**: filter all queries to `sessions.party_id = <party_id>`. Reflects how the group plays together.
- **Global user stats**: no party filter, join across all sessions the user participated in. Reflects the user's overall record regardless of group.

Both use the same underlying queries — party stats are just scoped.

---

## Stat Derivations (computed at query time)

| Stat | Scope | Derivation |
|---|---|---|
| Win rate | Party or global | Sessions where rank=1 or result='win' / total sessions (optionally filtered by party) |
| Head-to-head | Party or global | Sessions where both users participated, filtered by party if in party context |
| Win streak | Global | Consecutive wins across all parties, ordered by played_at |
| Nemesis | Party or global | Most common opponent in sessions where this user did not win |
| Sessions per month | Party or global | Group sessions by month of played_at |
| Most played game | Party or global | Game with highest session count for a user |
