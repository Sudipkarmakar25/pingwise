import { callBackend, toClientResponse, unavailableResponse } from "@/lib/server/backend";
import { clearSession, getSessionToken } from "@/lib/server/session";

export async function POST(request: Request) {
  const token = await getSessionToken();
  try {
    const upstream = await callBackend({
      method: "POST",
      path: "/api/v1/auth/change_password/",
      token,
      body: await request.text(),
    });
    if (upstream.status === 401 && token) {
      await clearSession();
    }
    return toClientResponse(upstream);
  } catch {
    return unavailableResponse();
  }
}
