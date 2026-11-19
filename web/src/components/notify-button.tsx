"use client";

import { useState } from "react";
import { useMe } from "@/components/app-shell";
import { api, ApiError } from "@/lib/api";
import type { MemberAlert } from "@/lib/types";

// NotifyButton shows whether the signed-in user gets alert emails for a service and lets them
// turn it on or off.
// compact is the smaller size used on service cards.
export function NotifyButton({
  serviceId,
  notify,
  onChange,
  compact = false,
}: {
  serviceId: string;
  notify: boolean;
  onChange: (notify: boolean) => void;
  compact?: boolean;
}) {
  const { me } = useMe();
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const on = notify;
  const label = on ? "Notifying" : "Notify";
  const title = error ?? (on ? "You get alert emails for this service. Click to turn off." : "Click to get alert emails for this service");
  const tone = on ? "bg-accent-soft text-accent-strong" : "bg-subtle text-muted";
  const shape = `gap-2 text-sm font-medium ${compact ? "h-8 px-3" : "h-10 px-4"}`;

  async function toggle() {
    if (!me) return;
    setSaving(true);
    setError(null);
    try {
      const res = await api<MemberAlert>(`/services/${serviceId}/alerts/${me.user.id}`, {
        method: "PATCH",
        body: { enabled: !notify },
      });
      onChange(res.enabled);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Couldn't change it. Try again.");
    } finally {
      setSaving(false);
    }
  }

  const content = (
    <>
      <BellIcon off={!on} />
      {label}
    </>
  );

  return (
    <button
      type="button"
      onClick={toggle}
      disabled={!me || saving}
      aria-pressed={on}
      title={title}
      className={`inline-flex items-center justify-center rounded-md transition-colors disabled:cursor-wait disabled:opacity-60 ${shape} ${tone} ${
        on ? "hover:bg-accent-soft-hover" : "hover:bg-subtle-hover hover:text-ink"
      } ${error ? "ring-1 ring-down" : ""}`}
    >
      {content}
    </button>
  );
}

function BellIcon({ off }: { off: boolean }) {
  return (
    <svg viewBox="0 0 24 24" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
      <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0" />
      {off && <path d="M3 3l18 18" />}
    </svg>
  );
}
