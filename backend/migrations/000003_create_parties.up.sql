CREATE TABLE parties (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    admin_user_id UUID NOT NULL REFERENCES users(id),
    invite_code   VARCHAR(32) UNIQUE NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_parties_admin       ON parties(admin_user_id);
CREATE INDEX idx_parties_invite_code ON parties(invite_code);
