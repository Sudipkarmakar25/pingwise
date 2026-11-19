import type { NextRequest } from "next/server";
import { callBackend, errorResponse, toClientResponse, unavailableResponse } from "@/lib/server/backend";
import { clearSession, getSessionToken } from "@/lib/server/session";

// Forwards /api/<path> from the browser to /api/v1/<path> on the Go API, adding the access
// token from the HTTP-only cookie. Auth endpoints have their own route handlers, because they
// set or clear the cookie and must never pass a token through to the browser.
async function forward(request: NextRequest, ctx: RouteContext<"/api/[...path]">) {
  const { path } = await ctx.params;
  if (path[0] === "auth" || path.some((segment) => segment === "." || segment === "..")) {
    return errorResponse(404, "Not found");
  }

  const token = await getSessionToken();
  const hasBody = request.method !== "GET" && request.method !== "HEAD";
  try {
    const upstream = await callBackend({
      method: request.method,
      path: `/api/v1/${path.map(encodeURIComponent).join("/")}/${request.nextUrl.search}`,
      token,
      body: hasBody ? await request.text() : undefined,
    });
    // the token was rejected (expired, signed out elsewhere, or user removed): drop the cookie,
    // otherwise the sign-in page would redirect straight back into the app
    if (upstream.status === 401 && token) {
      await clearSession();
    }
    return toClientResponse(upstream);
  } catch {
    return unavailableResponse();
  }
}

export { forward as GET, forward as POST, forward as PATCH, forward as DELETE };
