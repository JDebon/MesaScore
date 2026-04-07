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

CREATE INDEX idx_users_email              ON users(email);
CREATE INDEX idx_users_username           ON users(username);
CREATE INDEX idx_users_verification_token ON users(verification_token)
    WHERE verification_token IS NOT NULL;
