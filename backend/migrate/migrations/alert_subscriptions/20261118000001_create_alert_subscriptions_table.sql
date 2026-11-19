-- migrate:up
-- whether a member gets alert emails for a service. Without a row, owners get them and
-- editors and viewers do not.
CREATE TABLE IF NOT EXISTS alert_subscriptions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID        NOT NULL REFERENCES organisations (id),
    user_id    UUID        NOT NULL REFERENCES users (id),
    service_id UUID        NOT NULL REFERENCES services (id),
    is_enabled BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_by UUID        NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, service_id)
);

CREATE INDEX IF NOT EXISTS idx_alert_subscriptions_service_id ON alert_subscriptions (service_id) WHERE is_enabled;

-- migrate:down
DROP TABLE IF EXISTS alert_subscriptions;
