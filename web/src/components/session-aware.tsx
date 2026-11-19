import { cookies } from "next/headers";
import { Suspense } from "react";
import type { ReactNode } from "react";
import { SESSION_COOKIE } from "@/lib/server/config";

async function SessionSwitch({ signedIn, signedOut }: { signedIn: ReactNode; signedOut: ReactNode }) {
  const hasSession = (await cookies()).has(SESSION_COOKIE);
  return hasSession ? signedIn : signedOut;
}

// SessionAware shows different content to signed-in visitors on public pages. Only the cookie's
// presence is checked: an expired session reaches the app and is sent to sign in from there.
// The rest of the page stays static; this part fills in per request, keeping its space meanwhile.
export function SessionAware({ signedIn, signedOut }: { signedIn: ReactNode; signedOut: ReactNode }) {
  return (
    <Suspense fallback={<div className="invisible">{signedOut}</div>}>
      <SessionSwitch signedIn={signedIn} signedOut={signedOut} />
    </Suspense>
  );
}
