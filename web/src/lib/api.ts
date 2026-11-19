// Browser-side client for the Next.js /api routes. Every call goes to the same origin; the
// access token travels in the HTTP-only cookie and is added by the server.

export type ValidationError = { field: string; message: string };

export class ApiError extends Error {
  status: number;
  validationErrors: ValidationError[];

  constructor(status: number, message: string, validationErrors: ValidationError[] = []) {
    super(message);
    this.status = status;
    this.validationErrors = validationErrors;
  }

  // fieldError returns the message for one form field, if the API rejected it
  fieldError(field: string): string | undefined {
    return this.validationErrors.find((e) => e.field === field)?.message;
  }
}

type Options = { method?: "GET" | "POST" | "PATCH" | "DELETE"; body?: unknown };

export async function api<T>(path: string, { method = "GET", body }: Options = {}): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`/api${path}`, {
      method,
      headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      cache: "no-store",
    });
  } catch {
    throw new ApiError(0, "Can't reach Pingwise. Check your connection and try again.");
  }

  const payload = await response.json().catch(() => null);
  if (response.ok && payload?.success) {
    return payload.data as T;
  }

  // the session has expired or the user was removed: start again from the login page
  if (response.status === 401 && !path.startsWith("/auth/")) {
    // a full page load on purpose, so no state from the expired session survives
    // eslint-disable-next-line @next/next/no-location-assign-relative-destination
    window.location.assign(`/login?next=${encodeURIComponent(window.location.pathname)}`);
  }
  const data = payload?.data;
  throw new ApiError(
    response.status,
    data?.message || "Something went wrong. Try again.",
    Array.isArray(data?.validation_errors) ? data.validation_errors : [],
  );
}
