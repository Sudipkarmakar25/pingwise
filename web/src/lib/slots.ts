import { dayLabel, daysBetween } from "./format";
import type { DailyStat, Timeline } from "./types";

// A Slot is one bar in the uptime and response time charts: a minute, 5 minutes or a day
export type Slot = {
  key: string;
  label: string; // the slot's time, for tooltips
  total: number; // checks in the slot; 0 means no data
  up: number;
  avgMs: number;
  maxMs: number;
};

const timeFormat: Intl.DateTimeFormatOptions = { hour: "2-digit", minute: "2-digit" };

// slotsFromTimeline lists every slot in the window, filling slots without checks as empty
export function slotsFromTimeline(timeline: Timeline): Slot[] {
  const byStart = new Map(timeline.buckets.map((b) => [new Date(b.start).getTime(), b]));
  const step = timeline.bucket_minutes * 60_000;
  const slots: Slot[] = [];
  for (let t = new Date(timeline.from).getTime(); t <= new Date(timeline.to).getTime(); t += step) {
    const b = byStart.get(t);
    const start = new Date(t);
    const label =
      timeline.granularity === "day"
        ? dayLabel(start.toISOString().slice(0, 10))
        : timeline.granularity === "minute"
          ? `${start.toLocaleDateString("en-GB", { day: "numeric", month: "short" })}, ${start.toLocaleTimeString("en-GB", timeFormat)}`
          : `${start.toLocaleDateString("en-GB", { day: "numeric", month: "short" })}, ${start.toLocaleTimeString("en-GB", timeFormat)}–${new Date(t + step).toLocaleTimeString("en-GB", timeFormat)}`;
    slots.push({
      key: String(t),
      label,
      total: b?.total_checks ?? 0,
      up: b?.up_checks ?? 0,
      avgMs: b?.avg_response_ms ?? 0,
      maxMs: b?.max_response_ms ?? 0,
    });
  }
  return slots;
}

// slotsFromDailyStats builds daily slots for the dashboard, which loads stats for all services at once
export function slotsFromDailyStats(fromDay: string, toDay: string, stats: DailyStat[]): Slot[] {
  const byDay = new Map(stats.map((s) => [s.day, s]));
  return daysBetween(fromDay, toDay).map((day) => {
    const s = byDay.get(day);
    return {
      key: day,
      label: dayLabel(day),
      total: s?.total_checks ?? 0,
      up: s?.up_checks ?? 0,
      avgMs: s?.avg_response_ms ?? 0,
      maxMs: s?.max_response_ms ?? 0,
    };
  });
}
