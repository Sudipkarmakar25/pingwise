import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const SESSION_COOKIE = "pw_session";
const APP_PATHS = ["/dashboard", "/services", "/users", "/account"];
const AUTH_PATHS = ["/login", "/signup"];

// Optimistic check on the cookie only. The Go API still validates the token on every call,
// and an expired or revoked token sends the user back to /login from the page.
export function proxy(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  const hasSession = request.cookies.has(SESSION_COOKIE);

  if (!hasSession && APP_PATHS.some((p) => pathname === p || pathname.startsWith(`${p}/`))) {
    const login = new URL("/login", request.url);
    login.searchParams.set("next", pathname + search);
    return NextResponse.redirect(login);
  }
  if (hasSession && AUTH_PATHS.includes(pathname)) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*", "/services/:path*", "/users/:path*", "/account/:path*", "/login", "/signup"],
};
