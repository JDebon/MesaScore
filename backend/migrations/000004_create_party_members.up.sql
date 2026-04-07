CREATE TABLE party_members (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id  UUID NOT NULL REFERENCES parties(id),
    user_id   UUID NOT NULL REFERENCES users(id),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (party_id, user_id)
);

CREATE INDEX idx_party_members_party ON party_members(party_id);
CREATE INDEX idx_party_members_user  ON party_members(user_id);
