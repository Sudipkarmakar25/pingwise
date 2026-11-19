-- migrate:up
CREATE TABLE IF NOT EXISTS check_results (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    service_id       UUID        NOT NULL REFERENCES services (id),
    org_id           UUID        NOT NULL REFERENCES organisations (id),
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_up            BOOLEAN     NOT NULL,
    status_code      INTEGER,
    response_time_ms INTEGER,
    error            TEXT
);

CREATE INDEX IF NOT EXISTS idx_check_results_service_id_checked_at ON check_results (service_id, checked_at DESC);

-- migrate:down
DROP TABLE IF EXISTS check_results;
