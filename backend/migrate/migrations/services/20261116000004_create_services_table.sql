-- migrate:up
CREATE TABLE IF NOT EXISTS services (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                UUID        NOT NULL REFERENCES organisations (id),
    name                  TEXT        NOT NULL,
    url                   TEXT        NOT NULL,
    method                TEXT        NOT NULL DEFAULT 'GET' CHECK (method IN ('GET', 'POST', 'HEAD')),
    auth_header_name      TEXT,
    auth_token_encrypted  BYTEA,
    auth_token_last4      TEXT,
    auth_key_version      INTEGER,
    custom_headers        JSONB       NOT NULL DEFAULT '{}',
    request_body          TEXT,
    expected_status       INTEGER     NOT NULL DEFAULT 200,
    interval_minutes      INTEGER     NOT NULL DEFAULT 1 CHECK (interval_minutes IN (1, 2, 3, 5, 10, 15, 30, 60)),
    timeout_seconds       INTEGER     NOT NULL DEFAULT 10,
    failure_threshold     INTEGER     NOT NULL DEFAULT 2 CHECK (failure_threshold BETWEEN 1 AND 5),
    consecutive_failures  INTEGER     NOT NULL DEFAULT 0,
    is_active             BOOLEAN     NOT NULL DEFAULT TRUE,
    current_status        TEXT        NOT NULL DEFAULT 'unknown' CHECK (current_status IN ('up', 'down', 'unknown')),
    last_scheduled_at     TIMESTAMPTZ,
    last_checked_at       TIMESTAMPTZ,
    last_response_time_ms INTEGER,
    created_by            UUID        NOT NULL REFERENCES users (id),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_services_org_id ON services (org_id) WHERE deleted_at IS NULL;
-- used by the scheduler to find services that are due
CREATE INDEX IF NOT EXISTS idx_services_last_scheduled_at ON services (last_scheduled_at) WHERE is_active AND deleted_at IS NULL;

-- migrate:down
DROP TABLE IF EXISTS services;
