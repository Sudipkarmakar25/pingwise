"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { useMe } from "@/components/app-shell";
import { NotifyButton } from "@/components/notify-button";
import { displayStatus, StatusBadge } from "@/components/status-badge";
import type { DisplayStatus } from "@/components/status-badge";
import { Button, ButtonLink } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { UptimeBars } from "@/components/uptime-bars";
import { slotsFromDailyStats } from "@/lib/slots";
import { api, ApiError } from "@/lib/api";
import { exactTime, milliseconds, percent, relativeTime } from "@/lib/format";
import type { OrgStats, Service } from "@/lib/types";

type Data = { services: Service[]; stats: OrgStats; loadedAt: string };

// Services that need attention come first
const statusOrder: Record<DisplayStatus, number> = { down: 0, unknown: 1, up: 2, paused: 3 };

export function DashboardView() {
  const { me } = useMe();
  const [data, setData] = useState<Data | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);

  // the dashboard shows what the last scheduled checks found; Refresh reloads it on request
  const load = useCallback(async () => {
    setRefreshing(true);
    try {
      const [services, stats] = await Promise.all([api<Service[]>("/services"), api<OrgStats>("/services/stats")]);
      setData({ services, stats, loadedAt: new Date().toISOString() });
      setError(null);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong. Try again.");
    } finally {
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    // the request starts here; state is set only when it resolves
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [load]);

  // a member turned alert emails on or off for one service
  const setNotify = useCallback((serviceId: string, notify: boolean) => {
    setData((current) =>
      current && { ...current, services: current.services.map((s) => (s.id === serviceId ? { ...s, notify } : s)) },
    );
  }, []);

  const canEdit = me?.user.role === "owner" || me?.user.role === "editor";
  const services = [...(data?.services ?? [])].sort(
    (a, b) =>
      statusOrder[displayStatus(a.current_status, a.is_active)] -
        statusOrder[displayStatus(b.current_status, b.is_active)] || a.name.localeCompare(b.name),
  );

  return (
    <div>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Services</h1>
          <p className="mt-1 text-sm text-muted">
            {data ? (
              <>
                Showing results of the latest scheduled checks. Updated{" "}
                <time dateTime={data.loadedAt} title={exactTime(data.loadedAt)}>
                  {new Date(data.loadedAt).toLocaleTimeString("en-GB")}
                </time>
                .
              </>
            ) : (
              "Loading the latest results…"
            )}
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={load} disabled={refreshing} aria-live="polite">
            <RefreshIcon spinning={refreshing} />
            {refreshing ? "Refreshing…" : "Refresh"}
          </Button>
          {canEdit && (
            <ButtonLink href="/services/new" variant="primary">
              Add service
            </ButtonLink>
          )}
        </div>
      </div>

      {error && (
        <div role="alert" className="mt-6 flex items-center justify-between gap-4 rounded-lg border border-down/30 bg-down-soft px-4 py-3 text-sm text-down-ink">
          <span>{data ? `Couldn't refresh. ${error}` : `Couldn't load services. ${error}`}</span>
          <Button size="sm" onClick={load} disabled={refreshing}>
            Try again
          </Button>
        </div>
      )}

      {!data && !error && <LoadingState />}

      {data && services.length === 0 && <EmptyState canEdit={canEdit} />}

      {data && services.length > 0 && (
        <>
          <Summary services={services} />
          <ul className="mt-4 space-y-3">
            {services.map((service) => (
              <ServiceRow
                key={service.id}
                service={service}
                stats={data.stats}
                now={new Date(data.loadedAt).getTime()}
                canEdit={canEdit}
                onNotifyChange={setNotify}
              />
            ))}
          </ul>
        </>
      )}
    </div>
  );
}

function Summary({ services }: { services: Service[] }) {
  const counts = { up: 0, down: 0, unknown: 0, paused: 0 };
  for (const s of services) counts[displayStatus(s.current_status, s.is_active)]++;
  const items: [string, number, string][] = [
    ["Down", counts.down, counts.down > 0 ? "text-down-ink" : ""],
    ["Up", counts.up, ""],
    ["Pending", counts.unknown, ""],
    ["Paused", counts.paused, ""],
  ];
  return (
    <dl className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
      {items.map(([label, count, tone]) => (
        <div key={label} className="rounded-lg border border-line bg-surface px-4 py-3">
          <dt className="text-sm text-muted">{label}</dt>
          <dd className={`mt-0.5 font-mono text-2xl font-medium ${tone}`}>{count}</dd>
        </div>
      ))}
    </dl>
  );
}

function ServiceRow({
  service,
  stats,
  now,
  canEdit,
  onNotifyChange,
}: {
  service: Service;
  stats: OrgStats;
  now: number;
  canEdit: boolean;
  onNotifyChange: (serviceId: string, notify: boolean) => void;
}) {
  const status = displayStatus(service.current_status, service.is_active);
  const summary = stats.services[service.id];
  return (
    <li className="group relative rounded-xl border border-line bg-surface p-4 transition hover:border-muted/50 hover:shadow-sm sm:p-5">
      <div className="flex flex-wrap items-start justify-between gap-x-6 gap-y-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-3">
            {/* the link covers the whole card, so clicking anywhere on it opens the service */}
            <Link
              href={`/services/${service.id}`}
              className="truncate font-semibold group-hover:underline after:absolute after:inset-0 after:rounded-xl after:content-['']"
            >
              {service.name}
            </Link>
            <StatusBadge status={status} />
          </div>
        </div>
        <dl className="grid w-full grid-cols-3 gap-x-6 text-sm sm:w-auto sm:text-right">
          <div>
            <dt className="text-xs text-muted">Last check</dt>
            <dd className="mt-0.5">
              {service.last_checked_at ? (
                <time dateTime={service.last_checked_at} title={exactTime(service.last_checked_at)}>
                  {relativeTime(service.last_checked_at, now)}
                </time>
              ) : (
                <span className="text-muted">Not yet</span>
              )}
            </dd>
          </div>
          <div>
            <dt className="text-xs text-muted">Response</dt>
            <dd className="mt-0.5 font-mono">{milliseconds(service.last_response_time_ms)}</dd>
          </div>
          <div>
            <dt className="text-xs text-muted">90-day uptime</dt>
            <dd className="mt-0.5 font-mono">{percent(summary?.uptime_percentage)}</dd>
          </div>
        </dl>
      </div>
      {/* above the card link, so each day's tooltip still shows on hover */}
      <div className="relative z-10 mt-4">
        <UptimeBars
          slots={slotsFromDailyStats(stats.from_day, stats.to_day, summary?.daily_stats ?? [])}
          startLabel="90 days ago"
          endLabel="Today (UTC)"
          height="h-9"
          showLabels={false}
        />
      </div>
      <div className="mt-3 flex flex-wrap items-center justify-between gap-x-4 gap-y-2 text-xs text-muted">
        {/* method, URL and interval are on the service page; the card only flags an ongoing outage */}
        <span>
          <span className="md:hidden">Last 30 days</span>
          <span className="hidden md:inline">Last 90 days</span>
          {service.is_active && service.current_status === "down" && service.consecutive_failures > 0 && (
            <span className="text-down-ink">
              {" "}
              · {service.consecutive_failures} failed {service.consecutive_failures === 1 ? "check" : "checks"} in a row
            </span>
          )}
        </span>
        <span className="relative z-10 flex shrink-0 items-center gap-2">
          <NotifyButton
            serviceId={service.id}
            notify={service.notify}
            onChange={(notify) => onNotifyChange(service.id, notify)}
            compact
          />
          {canEdit && (
            <ButtonLink href={`/services/${service.id}/edit`} variant="muted" size="sm">
              Edit
            </ButtonLink>
          )}
          <ButtonLink href={`/services/${service.id}`} variant="soft" size="sm" className="group-hover:bg-accent-soft-hover">
            <span className="sm:hidden">Details</span>
            <span className="hidden sm:inline">View details</span>
            <span aria-hidden="true" className="transition-transform group-hover:translate-x-0.5">
              →
            </span>
          </ButtonLink>
        </span>
      </div>
    </li>
  );
}

function EmptyState({ canEdit }: { canEdit: boolean }) {
  return (
    <div className="mt-8 rounded-xl border border-dashed border-line bg-surface px-6 py-14 text-center">
      <h2 className="font-semibold">Add your first service</h2>
      <p className="mx-auto mt-2 max-w-md text-sm text-muted">
        {canEdit
          ? "Add an endpoint and Pingwise will start checking it within a minute."
          : "No services have been added yet. Ask an owner or editor in your organisation to add one."}
      </p>
      {canEdit && (
        <ButtonLink href="/services/new" variant="primary" className="mt-6">
          Add service
        </ButtonLink>
      )}
    </div>
  );
}

function LoadingState() {
  return (
    <div aria-busy="true" aria-label="Loading services">
      <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[0, 1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-[74px] rounded-lg" />
        ))}
      </div>
      <div className="mt-4 space-y-3">
        {[0, 1, 2].map((i) => (
          <Skeleton key={i} className="h-36 rounded-xl" />
        ))}
      </div>
    </div>
  );
}

function RefreshIcon({ spinning }: { spinning: boolean }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={`h-4 w-4 ${spinning ? "animate-spin" : ""}`}
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M21 12a9 9 0 1 1-2.64-6.36" />
      <path d="M21 3v6h-6" />
    </svg>
  );
}
