CREATE TABLE party_invites (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    party_id            UUID NOT NULL REFERENCES parties(id),
    invited_user_id     UUID NOT NULL REFERENCES users(id),
    invited_by_user_id  UUID NOT NULL REFERENCES users(id),
    status              invite_status NOT NULL DEFAULT 'pending',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_party_invites_party        ON party_invites(party_id);
CREATE INDEX idx_party_invites_invited_user ON party_invites(invited_user_id);
CREATE INDEX idx_party_invites_status       ON party_invites(invited_user_id, status)
    WHERE status = 'pending';

CREATE UNIQUE INDEX idx_party_invites_one_pending
    ON party_invites(party_id, invited_user_id)
    WHERE status = 'pending';
