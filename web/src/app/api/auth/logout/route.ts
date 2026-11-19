import { callBackend } from "@/lib/server/backend";
import { clearSession, getSessionToken } from "@/lib/server/session";

// Ends the session on the Go API and clears the cookie. The cookie is cleared even if the
// API call fails, so signing out always works from the browser's side.
export async function POST() {
  const token = await getSessionToken();
  if (token) {
    try {
      await callBackend({ method: "POST", path: "/api/v1/auth/logout/", token });
    } catch {
      // the session expires on its own; nothing more to do
    }
  }
  await clearSession();
  return Response.json({ success: true, data: null });
}
