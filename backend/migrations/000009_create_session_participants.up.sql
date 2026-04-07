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

CREATE INDEX idx_session_participants_session ON session_participants(session_id);
CREATE INDEX idx_session_participants_user    ON session_participants(user_id);
CREATE INDEX idx_session_participants_rank    ON session_participants(session_id, rank)
    WHERE rank IS NOT NULL;
