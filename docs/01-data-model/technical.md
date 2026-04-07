# Data Model — Technical Tasks

## Prerequisites
- PostgreSQL 16+
- Migration tool: `golang-migrate` (file-based, runs on app startup)
- All migrations live in `backend/migrations/`

---

## Tasks

### T1.1 — Define PostgreSQL enums

```sql
CREATE TYPE session_type      AS ENUM ('competitive', 'team', 'cooperative', 'score');
CREATE TYPE participant_result AS ENUM ('win', 'loss', 'draw');
CREATE TYPE invite_status     AS ENUM ('pending', 'accepted', 'declined');
```

---

### T1.2 — Create `users` table

```sql
CREATE TABLE users (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username             VARCHAR(50)  UNIQUE NOT NULL,
    display_name         VARCHAR(100) NOT NULL,
    email                VARCHAR(255) UNIQUE NOT NULL,
    password_hash        VARCHAR(255) NOT NULL,
    email_verified       BOOLEAN      NOT NULL DEFAULT FALSE,
    verification_token   VARCHAR(255) UNIQUE,
    verification_sent_at TIMESTAMPTZ,
    avatar_url           VARCHAR(500),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_users_email              ON users(email);
CREATE INDEX idx_users_username           ON users(username);
CREATE INDEX idx_users_verification_token ON users(verification_token)
    WHERE verification_token IS NOT NULL;
```

---

### T1.3 — Create `parties` table

```sql
CREATE TABLE parties (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    admin_user_id UUID NOT NULL REFERENCES users(id),
    invite_code   VARCHAR(32) UNIQUE NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_parties_admin       ON parties(admin_user_id);
CREATE INDEX idx_parties_invite_code ON parties(invite_code);
```

---

### T1.4 — Create `party_members` table

```sql
CREATE TABLE party_members (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id  UUID NOT NULL REFERENCES parties(id),
    user_id   UUID NOT NULL REFERENCES users(id),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (party_id, user_id)
);
```

**Indexes:**
```sql
CREATE INDEX idx_party_members_party ON party_members(party_id);
CREATE INDEX idx_party_members_user  ON party_members(user_id);
```

---

### T1.5 — Create `party_invites` table

```sql
CREATE TABLE party_invites (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id            UUID NOT NULL REFERENCES parties(id),
    invited_user_id     UUID NOT NULL REFERENCES users(id),
    invited_by_user_id  UUID NOT NULL REFERENCES users(id),
    status              invite_status NOT NULL DEFAULT 'pending',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_party_invites_party        ON party_invites(party_id);
CREATE INDEX idx_party_invites_invited_user ON party_invites(invited_user_id);
CREATE INDEX idx_party_invites_status       ON party_invites(invited_user_id, status)
    WHERE status = 'pending';
```

**Unique constraint for pending invites (partial index):**
```sql
CREATE UNIQUE INDEX idx_party_invites_one_pending
    ON party_invites(party_id, invited_user_id)
    WHERE status = 'pending';
```

---

### T1.6 — Create `games` table

```sql
CREATE TABLE games (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bgg_id            INTEGER UNIQUE,
    name              VARCHAR(255) NOT NULL,
    description       TEXT,
    cover_image_url   VARCHAR(500),
    min_players       INTEGER,
    max_players       INTEGER,
    bgg_rating        DECIMAL(4,2),
    bgg_fetched_at    TIMESTAMPTZ,
    added_by_user_id  UUID NOT NULL REFERENCES users(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Indexes:**
```sql
CREATE INDEX idx_games_bgg_id       ON games(bgg_id) WHERE bgg_id IS NOT NULL;
CREATE INDEX idx_games_name         ON games(name);
CREATE INDEX idx_games_added_by     ON games(added_by_user_id);
```

---

### T1.7 — Create `user_games` table

```sql
CREATE TABLE user_games (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id   UUID NOT NULL REFERENCES users(id),
    game_id   UUID NOT NULL REFERENCES games(id),
    added_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, game_id)
);
```

**Indexes:**
```sql
CREATE INDEX idx_user_games_user ON user_games(user_id);
CREATE INDEX idx_user_games_game ON user_games(game_id);
```

---

### T1.8 — Create `sessions` table

```sql
CREATE TABLE sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id            UUID NOT NULL REFERENCES parties(id),
    game_id             UUID NOT NULL REFERENCES games(id),
    session_type        session_type NOT NULL,
    played_at           TIMESTAMPTZ NOT NULL,
    duration_minutes    INTEGER,
    brought_by_user_id  UUID REFERENCES users(id),
    notes               TEXT,
    created_by          UUID NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT played_at_not_future CHECK (played_at <= NOW() + INTERVAL '1 hour')
);
```

**Indexes:**
```sql
CREATE INDEX idx_sessions_party_id   ON sessions(party_id);
CREATE INDEX idx_sessions_game_id    ON sessions(game_id);
CREATE INDEX idx_sessions_played_at  ON sessions(played_at DESC);
CREATE INDEX idx_sessions_created_by ON sessions(created_by);
CREATE INDEX idx_sessions_brought_by ON sessions(brought_by_user_id)
    WHERE brought_by_user_id IS NOT NULL;
```

---

### T1.9 — Create `session_participants` table

```sql
CREATE TABLE session_participants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id),
    team_name  VARCHAR(100),
    rank       INTEGER,
    score      DECIMAL(10,2),
    result     participant_result,
    UNIQUE (session_id, user_id),
    CONSTRAINT rank_positive CHECK (rank IS NULL OR rank > 0)
);
```

**Indexes:**
```sql
CREATE INDEX idx_session_participants_session ON session_participants(session_id);
CREATE INDEX idx_session_participants_user    ON session_participants(user_id);
CREATE INDEX idx_session_participants_rank    ON session_participants(session_id, rank)
    WHERE rank IS NOT NULL;
```

---

### T1.10 — Migration file structure

```
backend/migrations/
  000001_create_enums.up.sql / .down.sql
  000002_create_users.up.sql / .down.sql
  000003_create_parties.up.sql / .down.sql
  000004_create_party_members.up.sql / .down.sql
  000005_create_party_invites.up.sql / .down.sql
  000006_create_games.up.sql / .down.sql
  000007_create_user_games.up.sql / .down.sql
  000008_create_sessions.up.sql / .down.sql
  000009_create_session_participants.up.sql / .down.sql
```

---

### T1.11 — Seed data (development only)

`backend/scripts/seed.go`:
- 6 users (all verified), including 2 who are admins of different parties.
- 2 parties, each with 4-5 members.
- 1 user who belongs to both parties.
- 15 games in catalog, distributed across user collections.
- 5 sessions per party (mix of all session types).

---

### T1.12 — Invite code generation

Invite codes are 32-character URL-safe random strings. Generate with:

```go
func generateInviteCode() string {
    b := make([]byte, 24)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)[:32]
}
```

Generate on party creation and on each regeneration request. Store in `parties.invite_code`.

---

## Validation rules enforced at the application layer

| Rule | Where enforced |
|---|---|
| Email not yet verified → cannot log in | Login handler |
| Verification token expired (>24h) → reject | Verify email handler |
| Party admin cannot leave without transferring ownership | Leave party handler |
| `brought_by_user_id` must be a party member AND own the game | Session creation handler |
| All session participants must be party members | Session creation handler |
| At least 2 participants per session | Session creation handler |
| Session type field rules (rank/score/result/team_name) | Session creation handler |
| Only party admin can create/edit/delete sessions | Party membership + role check middleware |
| Only party admin can send invites | Party role check |
| Cannot invite a user who is already a member | Invite handler |
| Cannot invite a user who has a pending invite | Invite handler (enforced by partial unique index) |
| Ownership transfer target must be a current party member | Transfer handler |
