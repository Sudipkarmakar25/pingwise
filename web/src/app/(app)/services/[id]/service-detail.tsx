"use client";

import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import type { ReactNode } from "react";
import { useMe } from "@/components/app-shell";
import { BackLink } from "@/components/back-link";
import { NotifyButton } from "@/components/notify-button";
import { ResponseTimeChart } from "@/components/response-time-chart";
import { displayStatus, StatusBadge } from "@/components/status-badge";
import { Button, ButtonLink } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { UptimeBars } from "@/components/uptime-bars";
import { api, ApiError } from "@/lib/api";
import { checkErrorLabel } from "@/lib/check-errors";
import { exactTime, intervalLabel, milliseconds, percent, relativeTime } from "@/lib/format";
import { UPTIME_STOPS } from "@/lib/uptime-colour";
import { slotsFromTimeline } from "@/lib/slots";
import type { CheckPage, CheckResult, Granularity, MemberAlert, ServiceDetail, Timeline } from "@/lib/types";

const CHECKS_PAGE_SIZE = 25;

type Load =
  | { state: "loading" }
  | { state: "error"; status: number; message: string }
  | { state: "ready"; detail: ServiceDetail; checks: CheckResult[]; nextCursor: string; loadedAt: string };

export function ServiceDetailView() {
  const { id } = useParams<{ id: string }>();
  const { me } = useMe();
  const [load, setLoad] = useState<Load>({ state: "loading" });
  const [refreshing, setRefreshing] = useState(false);
  const [refreshError, setRefreshError] = useState<string | null>(null);

  const fetchAll = useCallback(async () => {
    setRefreshing(true);
    try {
      const [detail, page] = await Promise.all([
        api<ServiceDetail>(`/services/${id}`),
        api<CheckPage>(`/services/${id}/checks?limit=${CHECKS_PAGE_SIZE}`),
      ]);
      setLoad({
        state: "ready",
        detail,
        checks: page.checks,
        nextCursor: page.next_cursor,
        loadedAt: new Date().toISOString(),
      });
      setRefreshError(null);
    } catch (err) {
      const e = err instanceof ApiError ? err : new ApiError(0, "Something went wrong. Try again.");
      // keep what is on screen if a refresh fails
      setLoad((current) => (current.state === "ready" ? current : { state: "error", status: e.status, message: e.message }));
      setRefreshError(e.message);
    } finally {
      setRefreshing(false);
    }
  }, [id]);

  useEffect(() => {
    // the request starts here; state is set only when it resolves
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchAll();
  }, [fetchAll]);

  // the signed-in user's own alert setting changed, from the header button or the list
  const setNotify = useCallback((notify: boolean) => {
    setLoad((current) =>
      current.state === "ready"
        ? { ...current, detail: { ...current.detail, service: { ...current.detail.service, notify } } }
        : current,
    );
  }, []);

  const name = load.state === "ready" ? load.detail.service.name : null;
  useEffect(() => {
    if (name) document.title = `${name} · Pingwise`;
  }, [name]);

  if (load.state === "loading") return <LoadingState />;

  if (load.state === "error") {
    const notFound = load.status === 404 || load.status === 400;
    return (
      <div className="mx-auto max-w-xl rounded-xl border border-line bg-surface px-6 py-12 text-center">
        <h1 className="font-semibold">{notFound ? "This service doesn't exist" : "Couldn't load this service"}</h1>
        <p className="mx-auto mt-2 max-w-md text-sm text-muted">
          {notFound ? "It may have been deleted, or the link may be out of date." : load.message}
        </p>
        <div className="mt-6 flex justify-center gap-2">
          {!notFound && <Button onClick={fetchAll}>Try again</Button>}
          <ButtonLink href="/dashboard">Back to services</ButtonLink>
        </div>
      </div>
    );
  }

  const { detail, checks, nextCursor, loadedAt } = load;
  const { service } = detail;
  const canEdit = me?.user.role === "owner" || me?.user.role === "editor";
  const now = new Date(loadedAt).getTime();

  return (
    <div>
      <BackLink href="/dashboard">Services</BackLink>
      <div className="mt-3 flex flex-wrap items-end justify-between gap-4">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-semibold tracking-tight">{service.name}</h1>
            <StatusBadge status={displayStatus(service.current_status, service.is_active)} />
          </div>
          <p className="mt-1 truncate font-mono text-sm text-muted">
            {service.method} {service.url}
          </p>
        </div>
        <div className="flex gap-2">
          <NotifyButton serviceId={service.id} notify={service.notify} onChange={setNotify} />
          <Button onClick={fetchAll} disabled={refreshing}>
            {refreshing ? "Refreshing…" : "Refresh"}
          </Button>
          {canEdit && (
            <ButtonLink href={`/services/${service.id}/edit`} variant="primary">
              Edit
            </ButtonLink>
          )}
        </div>
      </div>

      {refreshError && (
        <div role="alert" className="mt-6 rounded-lg border border-down/30 bg-down-soft px-4 py-3 text-sm text-down-ink">
          Couldn&apos;t refresh. {refreshError}
        </div>
      )}
      {!service.is_active && (
        <div className="mt-6 rounded-lg border border-line bg-surface px-4 py-3 text-sm text-muted">
          Monitoring is paused. No checks run until it&apos;s resumed{canEdit ? " from the edit page" : ""}.
        </div>
      )}

      <dl className="mt-6 grid grid-cols-2 gap-3 lg:grid-cols-4">
        <Stat label="Last check">
          {service.last_checked_at ? (
            <time dateTime={service.last_checked_at} title={exactTime(service.last_checked_at)}>
              {relativeTime(service.last_checked_at, now)}
            </time>
          ) : (
            "Not yet"
          )}
        </Stat>
        <Stat label="Last response" mono>
          {milliseconds(service.last_response_time_ms)}
        </Stat>
        <Stat label="90-day uptime" mono>
          {percent(detail.uptime_percentage)}
        </Stat>
        <Stat label="Failures in a row" mono tone={service.consecutive_failures > 0 ? "text-down-ink" : ""}>
          {service.consecutive_failures} <span className="font-sans text-sm text-muted">of {service.failure_threshold}</span>
        </Stat>
      </dl>

      <History serviceId={service.id} refreshKey={loadedAt} />

      <div className="mt-6 grid gap-6 lg:grid-cols-[1fr_20rem]">
        <RecentChecks serviceId={service.id} initial={checks} initialCursor={nextCursor} now={now} />
        <div className="space-y-6">
          <Configuration detail={detail} />
          <AlertRecipients
            serviceId={service.id}
            refreshKey={`${loadedAt}-${service.notify}`}
            myId={me?.user.id}
            onMyChange={setNotify}
          />
        </div>
      </div>
    </div>
  );
}

const GRANULARITIES: { value: Granularity; label: string; unit: string; start: string; end: string }[] = [
  { value: "minute", label: "Per minute · last 3 hours", unit: "minute", start: "3 hours ago", end: "Now" },
  { value: "five_minutes", label: "Every 5 minutes · last 24 hours", unit: "5 minutes", start: "24 hours ago", end: "Now" },
  { value: "day", label: "Daily · last 90 days", unit: "day (UTC)", start: "90 days ago", end: "Today (UTC)" },
];

// History shows uptime and response time for the chosen granularity. One dropdown drives both charts.
function History({ serviceId, refreshKey }: { serviceId: string; refreshKey: string }) {
  const [granularity, setGranularity] = useState<Granularity>("five_minutes");
  const [timeline, setTimeline] = useState<Timeline | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const option = GRANULARITIES.find((g) => g.value === granularity)!;

  const load = useCallback(async () => {
    setLoading(true);
    try {
      setTimeline(await api<Timeline>(`/services/${serviceId}/timeline?granularity=${granularity}`));
      setError(null);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong. Try again.");
    } finally {
      setLoading(false);
    }
  }, [serviceId, granularity]);

  useEffect(() => {
    // reload when the granularity changes or the page is refreshed
    // eslint-disable-next-line react-hooks/set-state-in-effect
    load();
  }, [load, refreshKey]);

  // only show data for the chosen granularity, never the previous one while the new one loads
  const current = timeline?.granularity === granularity ? timeline : null;
  const slots = current ? slotsFromTimeline(current) : [];
  const total = slots.reduce((sum, s) => sum + s.total, 0);
  const up = slots.reduce((sum, s) => sum + s.up, 0);
  const periodUptime = total > 0 ? Math.round((up / total) * 10000) / 100 : null;

  return (
    <section className="mt-6 rounded-xl border border-line bg-surface">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line px-5 py-4">
        <div>
          <h2 className="font-semibold">History</h2>
          <p className="mt-0.5 text-sm text-muted">One bar per {option.unit}. Hover a bar for its figures.</p>
        </div>
        <label className="flex items-center gap-2 text-sm">
          <span className="text-muted">Show</span>
          <select
            value={granularity}
            onChange={(e) => setGranularity(e.target.value as Granularity)}
            className="h-9 rounded-md border border-line bg-surface px-2.5 text-sm"
          >
            {GRANULARITIES.map((g) => (
              <option key={g.value} value={g.value}>
                {g.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      {error && (
        <div role="alert" className="mx-5 mt-4 rounded-lg border border-down/30 bg-down-soft px-4 py-3 text-sm text-down-ink">
          Couldn&apos;t load history. {error}
        </div>
      )}

      <div className="space-y-8 p-5" aria-busy={loading}>
        <div>
          <div className="mb-3 flex items-baseline justify-between gap-4">
            <h3 className="text-sm font-semibold">Uptime</h3>
            {current && (
              <span className="text-sm text-muted">
                <span className="font-mono text-ink">{percent(periodUptime)}</span> over this period
              </span>
            )}
          </div>
          {current ? (
            <UptimeBars slots={slots} startLabel={option.start} endLabel={option.end} height="h-10" />
          ) : (
            <Skeleton className="h-[62px]" />
          )}
          <Legend />
        </div>

        <div>
          <h3 className="mb-3 text-sm font-semibold">Response time</h3>
          {!current ? (
            <Skeleton className="h-[134px]" />
          ) : total > 0 ? (
            <ResponseTimeChart slots={slots} startLabel={option.start} endLabel={option.end} />
          ) : (
            <p className="text-sm text-muted">No checks in this period.</p>
          )}
        </div>
      </div>
    </section>
  );
}

function RecentChecks({
  serviceId,
  initial,
  initialCursor,
  now,
}: {
  serviceId: string;
  initial: CheckResult[];
  initialCursor: string;
  now: number;
}) {
  const [extra, setExtra] = useState<{ checks: CheckResult[]; cursor: string; forInitial: CheckResult[] }>({
    checks: [],
    cursor: initialCursor,
    forInitial: initial,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // a refresh replaces the first page, so drop any older pages loaded for the previous one
  const current = extra.forInitial === initial ? extra : { checks: [], cursor: initialCursor, forInitial: initial };
  const checks = [...initial, ...current.checks];

  async function loadMore() {
    setLoading(true);
    setError(null);
    try {
      const page = await api<CheckPage>(
        `/services/${serviceId}/checks?limit=${CHECKS_PAGE_SIZE}&cursor=${encodeURIComponent(current.cursor)}`,
      );
      setExtra({ checks: [...current.checks, ...page.checks], cursor: page.next_cursor, forInitial: initial });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong. Try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <section className="rounded-xl border border-line bg-surface">
      <div className="border-b border-line px-5 py-4">
        <h2 className="font-semibold">Recent checks</h2>
        <p className="mt-0.5 text-sm text-muted">Every check, newest first. Times are in your timezone.</p>
      </div>
      {checks.length === 0 ? (
        <p className="px-5 py-10 text-center text-sm text-muted">
          No checks yet. The first check runs within a minute of adding a service.
        </p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-line text-left text-xs text-muted">
                <th scope="col" className="px-5 py-2.5 font-medium">Time</th>
                <th scope="col" className="px-3 py-2.5 font-medium">Result</th>
                <th scope="col" className="px-3 py-2.5 font-medium">Status</th>
                <th scope="col" className="px-3 py-2.5 text-right font-medium">Response</th>
                <th scope="col" className="px-5 py-2.5 font-medium">Reason</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {checks.map((check) => (
                <tr key={check.id}>
                  <td className="px-5 py-2.5 whitespace-nowrap">
                    <time dateTime={check.checked_at} title={relativeTime(check.checked_at, now)}>
                      {new Date(check.checked_at).toLocaleString("en-GB", {
                        day: "numeric",
                        month: "short",
                        hour: "2-digit",
                        minute: "2-digit",
                        second: "2-digit",
                      })}
                    </time>
                  </td>
                  <td className="px-3 py-2.5">
                    <span
                      className={`inline-flex items-center gap-1.5 text-xs font-medium ${check.is_up ? "text-up-ink" : "text-down-ink"}`}
                    >
                      <span className={`h-1.5 w-1.5 rounded-full ${check.is_up ? "bg-up" : "bg-down"}`} aria-hidden="true" />
                      {check.is_up ? "Passed" : "Failed"}
                    </span>
                  </td>
                  <td className="px-3 py-2.5 font-mono">{check.status_code ?? "—"}</td>
                  <td className="px-3 py-2.5 text-right font-mono whitespace-nowrap">
                    {milliseconds(check.response_time_ms)}
                  </td>
                  <td className="px-5 py-2.5 text-muted">{checkErrorLabel(check.error) ?? "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {(current.cursor || error) && (
        <div className="flex items-center gap-3 border-t border-line px-5 py-3">
          {current.cursor && (
            <Button size="sm" onClick={loadMore} disabled={loading}>
              {loading ? "Loading…" : "Load older checks"}
            </Button>
          )}
          {error && (
            <span role="alert" className="text-sm text-down-ink">
              {error}
            </span>
          )}
        </div>
      )}
    </section>
  );
}

function Configuration({ detail }: { detail: ServiceDetail }) {
  const { service } = detail;
  const headers = Object.keys(service.custom_headers);
  const rows: [string, ReactNode][] = [
    ["Method", <span key="m" className="font-mono">{service.method}</span>],
    ["Expected status", <span key="s" className="font-mono">{service.expected_status}</span>],
    ["Timeout", `${service.timeout_seconds} seconds`],
    ["Interval", intervalLabel(service.interval_minutes)],
    ["Marked down after", `${service.failure_threshold} failed ${service.failure_threshold === 1 ? "check" : "checks in a row"}`],
    [
      "Authentication",
      service.auth_header_name ? (
        <span key="a" className="font-mono">
          {service.auth_header_name}: {service.auth_token_masked}
        </span>
      ) : (
        "None"
      ),
    ],
    ["Custom headers", headers.length ? <span key="h" className="font-mono">{headers.join(", ")}</span> : "None"],
  ];
  if (service.method === "POST") rows.push(["Request body", service.request_body ? "Set" : "None"]);

  return (
    <section className="h-fit rounded-xl border border-line bg-surface">
      <div className="border-b border-line px-5 py-4">
        <h2 className="font-semibold">Configuration</h2>
      </div>
      <dl className="divide-y divide-line text-sm">
        {rows.map(([label, value]) => (
          <div key={label} className="flex justify-between gap-4 px-5 py-2.5">
            <dt className="text-muted">{label}</dt>
            <dd className="min-w-0 text-right break-words">{value}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}

const roleLabel = { owner: "Owner", editor: "Editor", viewer: "Viewer" } as const;

// AlertRecipients lists who gets alert emails for the service. Owners get them by default, editors
// and viewers do not. Members can turn their own on or off, and a higher role can change a lower role's.
function AlertRecipients({
  serviceId,
  refreshKey,
  myId,
  onMyChange,
}: {
  serviceId: string;
  refreshKey: string;
  myId: string | undefined;
  onMyChange: (notify: boolean) => void;
}) {
  const [members, setMembers] = useState<MemberAlert[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [savingId, setSavingId] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    api<MemberAlert[]>(`/services/${serviceId}/alerts`)
      .then((res) => {
        if (cancelled) return;
        setMembers(res);
        setError(null);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : "Couldn't load who gets alerts.");
      });
    return () => {
      cancelled = true;
    };
  }, [serviceId, refreshKey]);

  async function change(member: MemberAlert, enabled: boolean) {
    setSavingId(member.user_id);
    setError(null);
    try {
      const updated = await api<MemberAlert>(`/services/${serviceId}/alerts/${member.user_id}`, {
        method: "PATCH",
        body: { enabled },
      });
      setMembers((current) => current?.map((m) => (m.user_id === updated.user_id ? updated : m)) ?? null);
      if (updated.user_id === myId) onMyChange(updated.enabled);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Couldn't change it. Try again.");
    } finally {
      setSavingId(null);
    }
  }

  const count = members?.filter((m) => m.enabled).length ?? 0;

  return (
    <section className="h-fit rounded-xl border border-line bg-surface">
      <div className="border-b border-line px-5 py-4">
        <h2 className="font-semibold">Alert emails</h2>
        <p className="mt-0.5 text-sm text-muted">
          {members ? `${count} ${count === 1 ? "person gets" : "people get"} an email when this service goes down or recovers.` : "Loading…"}
        </p>
      </div>
      {error && (
        <p role="alert" className="border-b border-line bg-down-soft px-5 py-2.5 text-sm text-down-ink">
          {error}
        </p>
      )}
      {members && (
        <ul className="divide-y divide-line text-sm">
          {members.map((m) => {
            const changedBySomeoneElse = m.updated_by_name && m.updated_by_name !== m.name;
            return (
              <li key={m.user_id} className="flex items-center justify-between gap-4 px-5 py-2.5">
                <div className="min-w-0">
                  <div className="truncate">
                    {m.name}
                    {m.user_id === myId && <span className="text-muted"> (you)</span>}
                  </div>
                  <div className="text-xs text-muted">
                    {roleLabel[m.role]}
                    {changedBySomeoneElse && ` · ${m.enabled ? "turned on" : "turned off"} by ${m.updated_by_name}`}
                  </div>
                </div>
                <Switch
                  checked={m.enabled}
                  onChange={(enabled) => change(m, enabled)}
                  disabled={!m.can_change || savingId === m.user_id}
                  label={`Alert emails for ${m.name}`}
                  title={m.can_change ? undefined : "Only they or someone with a higher role can change this"}
                />
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}

function Stat({ label, mono, tone = "", children }: { label: string; mono?: boolean; tone?: string; children: ReactNode }) {
  return (
    <div className="rounded-lg border border-line bg-surface px-4 py-3">
      <dt className="text-sm text-muted">{label}</dt>
      <dd className={`mt-0.5 text-xl font-medium ${mono ? "font-mono" : ""} ${tone}`}>{children}</dd>
    </div>
  );
}


// Legend shows the colour scale with its stops spaced evenly, so the narrow 95–100% range is readable
function Legend() {
  const gradient = `linear-gradient(to right, ${UPTIME_STOPS.map(([, colour], i) => `${colour} ${(i / (UPTIME_STOPS.length - 1)) * 100}%`).join(", ")})`;
  return (
    <div className="mt-4 flex flex-wrap items-end gap-x-6 gap-y-3 text-xs text-muted">
      <div className="w-full max-w-sm">
        <div className="h-2 rounded-full" style={{ background: gradient }} aria-hidden="true" />
        <div className="mt-1 flex justify-between font-mono">
          {UPTIME_STOPS.map(([p]) => (
            <span key={p}>{p}%</span>
          ))}
        </div>
      </div>
      <span className="flex items-center gap-1.5 pb-[18px]">
        <span className="h-2.5 w-2.5 rounded-sm bg-nodata" aria-hidden="true" />
        No data
      </span>
    </div>
  );
}

function LoadingState() {
  return (
    <div aria-busy="true" aria-label="Loading service" className="space-y-4">
      <Skeleton className="h-4 w-20" />
      <Skeleton className="h-8 w-72" />
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {[0, 1, 2, 3].map((i) => (
          <Skeleton key={i} className="h-[74px] rounded-lg" />
        ))}
      </div>
      <Skeleton className="h-36 rounded-xl" />
      <Skeleton className="h-48 rounded-xl" />
    </div>
  );
}
