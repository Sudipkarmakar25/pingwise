import { startSession } from "@/lib/server/session";

export async function POST(request: Request) {
  return startSession("/api/v1/auth/login/", request);
}
