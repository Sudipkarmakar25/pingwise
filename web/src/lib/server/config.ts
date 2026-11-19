import "server-only";

// Base URL of the Go API. Only the Next.js server talks to it; the browser never does.
export const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8090";

// HTTP-only cookie that holds the access token
export const SESSION_COOKIE = "pw_session";

// Same as the access token lifetime in the Go API
export const SESSION_MAX_AGE_SECONDS = 30 * 24 * 60 * 60;
