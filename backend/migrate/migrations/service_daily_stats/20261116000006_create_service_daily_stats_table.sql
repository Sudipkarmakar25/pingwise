-- migrate:up
CREATE TABLE IF NOT EXISTS service_daily_stats (
    service_id        UUID    NOT NULL REFERENCES services (id),
    org_id            UUID    NOT NULL REFERENCES organisations (id),
    day               DATE    NOT NULL,
    total_checks      INTEGER NOT NULL DEFAULT 0,
    up_checks         INTEGER NOT NULL DEFAULT 0,
    total_response_ms BIGINT  NOT NULL DEFAULT 0,
    max_response_ms   INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (service_id, day)
);

-- migrate:down
DROP TABLE IF EXISTS service_daily_stats;
