import type { ServiceStatus } from "@/lib/types";

export type DisplayStatus = ServiceStatus | "paused";

// displayStatus is what a service shows: a paused service shows "Paused" whatever its last result
export function displayStatus(status: ServiceStatus, isActive: boolean): DisplayStatus {
  return isActive ? status : "paused";
}

const styles: Record<DisplayStatus, { label: string; badge: string; dot: string }> = {
  up: { label: "Up", badge: "bg-up-soft text-up-ink", dot: "bg-up" },
  down: { label: "Down", badge: "bg-down-soft text-down-ink", dot: "bg-down" },
  unknown: { label: "Pending", badge: "bg-canvas text-muted ring-1 ring-line", dot: "bg-muted" },
  paused: { label: "Paused", badge: "bg-canvas text-muted ring-1 ring-line", dot: "bg-nodata" },
};

// StatusBadge shows status as a word and a colour, never colour alone
export function StatusBadge({ status }: { status: DisplayStatus }) {
  const style = styles[status];
  return (
    <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${style.badge}`}>
      <span className={`h-1.5 w-1.5 rounded-full ${style.dot}`} aria-hidden="true" />
      {style.label}
    </span>
  );
}
