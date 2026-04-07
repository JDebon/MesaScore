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

CREATE INDEX idx_sessions_party_id   ON sessions(party_id);
CREATE INDEX idx_sessions_game_id    ON sessions(game_id);
CREATE INDEX idx_sessions_played_at  ON sessions(played_at DESC);
CREATE INDEX idx_sessions_created_by ON sessions(created_by);
CREATE INDEX idx_sessions_brought_by ON sessions(brought_by_user_id)
    WHERE brought_by_user_id IS NOT NULL;
