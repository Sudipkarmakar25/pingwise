// Formatting helpers shared by the app pages

const relativeFormat = new Intl.RelativeTimeFormat("en", { numeric: "auto" });

// relativeTime turns a timestamp into "just now", "2 minutes ago", "3 hours ago" and so on
export function relativeTime(iso: string, now = Date.now()) {
  const seconds = Math.round((new Date(iso).getTime() - now) / 1000);
  const abs = Math.abs(seconds);
  if (abs < 45) return "just now";
  if (abs < 45 * 60) return relativeFormat.format(Math.round(seconds / 60), "minute");
  if (abs < 22 * 3600) return relativeFormat.format(Math.round(seconds / 3600), "hour");
  return relativeFormat.format(Math.round(seconds / 86400), "day");
}

// exactTime is the full date and time in the viewer's timezone, for tooltips
export function exactTime(iso: string) {
  return new Date(iso).toLocaleString("en-GB", {
    dateStyle: "medium",
    timeStyle: "medium",
  });
}

// dayLabel formats a UTC date like 2026-11-16 as "16 Nov 2026"
export function dayLabel(day: string) {
  return new Date(`${day}T00:00:00Z`).toLocaleDateString("en-GB", {
    day: "numeric",
    month: "short",
    year: "numeric",
    timeZone: "UTC",
  });
}

// daysBetween lists every UTC date from fromDay to toDay, inclusive, as YYYY-MM-DD
export function daysBetween(fromDay: string, toDay: string) {
  const days: string[] = [];
  const end = new Date(`${toDay}T00:00:00Z`);
  for (let d = new Date(`${fromDay}T00:00:00Z`); d <= end; d.setUTCDate(d.getUTCDate() + 1)) {
    days.push(d.toISOString().slice(0, 10));
  }
  return days;
}

export function intervalLabel(minutes: number) {
  if (minutes === 60) return "Every hour";
  return minutes === 1 ? "Every minute" : `Every ${minutes} minutes`;
}

export function percent(value: number | null | undefined) {
  if (value === null || value === undefined) return "—";
  return `${value.toFixed(value === 100 ? 0 : 2)}%`;
}

export function milliseconds(value: number | null | undefined) {
  if (value === null || value === undefined) return "—";
  return `${value.toLocaleString("en-GB")} ms`;
}
