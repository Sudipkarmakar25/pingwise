import { uptimeColour } from "@/lib/uptime-colour";

// Static picture of the dashboard for the home page. Illustrative data only.
// Each day is that day's uptime percentage, or null for days without checks.
type Day = number | null;

const services: { name: string; url: string; status: "up" | "down"; uptime: string; ms: number; days: Day[] }[] = [
  { name: "Public API", url: "api.example.com/health", status: "up", uptime: "99.98", ms: 142, days: pattern({ 17: 99.6 }) },
  { name: "Payments", url: "payments.example.com/status", status: "up", uptime: "99.71", ms: 231, days: pattern({ 9: 99.2, 22: 97.8, 23: 99.7 }) },
  {
    name: "Webhook relay",
    url: "hooks.example.com/ping",
    status: "down",
    uptime: "97.40",
    ms: 0,
    days: pattern({ 4: 98.4, 15: 96.1, 27: 99.4, 28: 91, 29: 62 }),
  },
];

// pattern makes 30 days at 100% except the given ones; the first two days have no data
function pattern(exceptions: Record<number, number>): Day[] {
  return Array.from({ length: 30 }, (_, i) => (i < 2 ? null : (exceptions[i] ?? 100)));
}

export function DashboardPreview() {
  return (
    <figure
      aria-label="Example of the Pingwise dashboard"
      className="rounded-xl border border-line bg-surface p-5 text-ink shadow-[0_24px_60px_-20px_rgba(4,20,40,0.55)]"
    >
      <div className="flex items-center justify-between">
        <div>
          <div className="text-sm font-semibold">Services</div>
          <div className="text-xs text-muted">Last refreshed a few seconds ago</div>
        </div>
        <span className="rounded-md border border-line px-2.5 py-1 text-xs font-medium">Refresh</span>
      </div>

      <ul className="mt-4 space-y-3">
        {services.map((s) => (
          <li key={s.name} className="rounded-lg border border-line p-3">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <div className="text-sm font-medium">{s.name}</div>
                <div className="truncate font-mono text-[11px] text-muted">{s.url}</div>
              </div>
              <span
                className={`shrink-0 rounded-full px-2 py-0.5 text-[11px] font-medium ${
                  s.status === "up" ? "bg-up-soft text-up-ink" : "bg-down-soft text-down-ink"
                }`}
              >
                {s.status === "up" ? "Up" : "Down"}
              </span>
            </div>
            <div className="mt-2.5 flex h-6 gap-[2px]">
              {s.days.map((d, i) => (
                <span
                  key={i}
                  className={`flex-1 rounded-[2px] ${d === null ? "bg-nodata" : ""}`}
                  style={d === null ? undefined : { backgroundColor: uptimeColour(d) }}
                />
              ))}
            </div>
            <div className="mt-1.5 flex justify-between text-[11px] text-muted">
              <span>30 days ago</span>
              <span className="font-mono">
                {s.uptime}% uptime{s.ms ? ` · ${s.ms} ms` : ""}
              </span>
              <span>Today</span>
            </div>
          </li>
        ))}
      </ul>
    </figure>
  );
}
