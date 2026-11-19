import "server-only";
import { API_BASE_URL } from "./config";

// Shape of every error response from the Go API, reused for errors raised here
export function errorResponse(status: number, message: string) {
  return Response.json(
    { success: false, data: { message, status_code: status, error_code: "", validation_errors: null } },
    { status },
  );
}

export const unavailableResponse = () =>
  errorResponse(503, "Pingwise is temporarily unavailable. Try again in a moment.");

type BackendRequest = {
  method: string;
  path: string; // path under the API origin, for example /api/v1/services/
  token?: string;
  body?: string;
};

// callBackend sends a request to the Go API. The origin is fixed, so neither the request
// nor the token can be pointed at another host.
export async function callBackend({ method, path, token, body }: BackendRequest): Promise<Response> {
  const headers: Record<string, string> = { Accept: "application/json" };
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (token) headers.Authorization = `Bearer ${token}`;

  return fetch(new URL(path, API_BASE_URL), {
    method,
    headers,
    body,
    redirect: "manual",
    cache: "no-store",
  });
}

// toClientResponse copies the backend's status and JSON body into a response for the browser
export async function toClientResponse(upstream: Response): Promise<Response> {
  const text = await upstream.text();
  return new Response(text, {
    status: upstream.status,
    headers: { "Content-Type": upstream.headers.get("Content-Type") ?? "application/json" },
  });
}
