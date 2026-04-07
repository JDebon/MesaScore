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

CREATE INDEX idx_games_bgg_id       ON games(bgg_id) WHERE bgg_id IS NOT NULL;
CREATE INDEX idx_games_name         ON games(name);
CREATE INDEX idx_games_added_by     ON games(added_by_user_id);
