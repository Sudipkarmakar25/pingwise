// safeNext returns the ?next= target only if it is a path on this site, so a crafted link
// can't send someone to another website after they sign in.
export function safeNext(next: string | null, fallback = "/dashboard") {
  if (!next || !next.startsWith("/") || next.startsWith("//") || next.startsWith("/\\")) {
    return fallback;
  }
  return next;
}
