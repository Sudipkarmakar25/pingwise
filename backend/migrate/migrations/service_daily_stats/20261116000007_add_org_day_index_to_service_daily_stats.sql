-- migrate:up
-- used by the dashboard, which reads the daily stats of every service in an org at once
CREATE INDEX IF NOT EXISTS idx_service_daily_stats_org_id_day ON service_daily_stats (org_id, day);

-- migrate:down
DROP INDEX IF EXISTS idx_service_daily_stats_org_id_day;
