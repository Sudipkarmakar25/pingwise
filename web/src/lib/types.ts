// Shapes returned by the Go API

export type Role = "owner" | "editor" | "viewer";

export type User = {
  id: string;
  org_id: string;
  email: string;
  name: string;
  role: Role;
  added_by: string | null;
  must_change_password: boolean;
  created_at: string;
  updated_at: string;
};

export type Organisation = {
  id: string;
  name: string;
  slug: string;
  created_at: string;
  updated_at: string;
};

export type Me = { user: User; organisation: Organisation };

export type ServiceStatus = "up" | "down" | "unknown";

export type Service = {
  id: string;
  org_id: string;
  name: string;
  url: string;
  method: "GET" | "POST" | "HEAD";
  auth_header_name: string | null;
  auth_token_masked: string | null;
  custom_headers: Record<string, string>;
  request_body: string | null;
  expected_status: number;
  interval_minutes: number;
  timeout_seconds: number;
  failure_threshold: number;
  consecutive_failures: number;
  is_active: boolean;
  current_status: ServiceStatus;
  last_checked_at: string | null;
  last_response_time_ms: number | null;
  created_by: string;
  created_at: string;
  updated_at: string;
  notify: boolean; // whether the signed-in user gets alert emails for this service
};

// One member's alert emails for a service, as the signed-in user sees it
export type MemberAlert = {
  user_id: string;
  name: string;
  role: Role;
  enabled: boolean;
  can_change: boolean; // whether the signed-in user may turn it on or off
  updated_by_name: string | null;
  updated_at: string | null;
};

export type DailyStat = {
  day: string; // UTC date, YYYY-MM-DD
  total_checks: number;
  up_checks: number;
  avg_response_ms: number;
  max_response_ms: number;
};

export type StatsSummary = {
  daily_stats: DailyStat[];
  uptime_percentage: number | null;
};

export type ServiceDetail = StatsSummary & {
  service: Service;
  from_day: string;
  to_day: string;
};

export type CheckResult = {
  id: number;
  service_id: string;
  checked_at: string;
  is_up: boolean;
  status_code: number | null;
  response_time_ms: number;
  error: string | null;
};

export type CheckPage = { checks: CheckResult[]; next_cursor: string };

export type Granularity = "minute" | "five_minutes" | "day";

export type TimelineBucket = {
  start: string; // RFC 3339, UTC
  total_checks: number;
  up_checks: number;
  avg_response_ms: number;
  max_response_ms: number;
};

export type Timeline = {
  granularity: Granularity;
  bucket_minutes: number;
  from: string;
  to: string;
  buckets: TimelineBucket[]; // only slots that had checks
};

export type OrgStats = {
  from_day: string;
  to_day: string;
  services: Record<string, StatsSummary>;
};
