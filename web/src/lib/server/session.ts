import "server-only";
import { cookies } from "next/headers";
import { callBackend, toClientResponse, unavailableResponse } from "./backend";
import { SESSION_COOKIE, SESSION_MAX_AGE_SECONDS } from "./config";

export async function getSessionToken() {
  return (await cookies()).get(SESSION_COOKIE)?.value;
}

// startSession calls signup or login on the Go API. On success it keeps the access token in an
// HTTP-only cookie and returns only the user to the browser; the tokens never leave the server.
export async function startSession(path: string, request: Request): Promise<Response> {
  let upstream: Response;
  try {
    upstream = await callBackend({ method: "POST", path, body: await request.text() });
  } catch {
    return unavailableResponse();
  }
  if (!upstream.ok) {
    return toClientResponse(upstream);
  }

  const { data } = await upstream.json();
  (await cookies()).set(SESSION_COOKIE, data.access_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: SESSION_MAX_AGE_SECONDS,
  });
  return Response.json({ success: true, data: { user: data.user } }, { status: upstream.status });
}

export async function clearSession() {
  (await cookies()).delete(SESSION_COOKIE);
}
