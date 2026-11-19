-- migrate:up
CREATE TABLE IF NOT EXISTS sessions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID        NOT NULL REFERENCES users (id),
    org_id             UUID        NOT NULL REFERENCES organisations (id),
    access_token_hash  TEXT        NOT NULL UNIQUE,
    refresh_token_hash TEXT        NOT NULL UNIQUE,
    access_expires_at  TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);

-- migrate:down
DROP TABLE IF EXISTS sessions;
