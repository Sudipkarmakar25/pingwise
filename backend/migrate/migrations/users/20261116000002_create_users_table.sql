-- migrate:up
CREATE TABLE IF NOT EXISTS users (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id               UUID        NOT NULL REFERENCES organisations (id),
    email                TEXT        NOT NULL,
    password_hash        TEXT        NOT NULL,
    name                 TEXT        NOT NULL,
    role                 TEXT        NOT NULL CHECK (role IN ('owner', 'editor', 'viewer')),
    added_by             UUID REFERENCES users (id),
    must_change_password BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ
);

-- a deleted user's email can be registered again as a new user
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_org_id ON users (org_id) WHERE deleted_at IS NULL;

-- migrate:down
DROP TABLE IF EXISTS users;
