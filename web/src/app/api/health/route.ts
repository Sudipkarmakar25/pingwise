import { callBackend } from "@/lib/server/backend";

// Health of the web app and of the Go API behind it
export async function GET() {
  let api = "unreachable";
  try {
    const upstream = await callBackend({ method: "GET", path: "/api/health/" });
    api = upstream.ok ? "ok" : `error ${upstream.status}`;
  } catch {
    // api stays unreachable
  }
  return Response.json({ web: "ok", api }, { status: api === "ok" ? 200 : 503 });
}
