import { milliseconds } from "@/lib/format";
import type { Slot } from "@/lib/slots";

const MOBILE_SLOTS = 30;

// ResponseTimeChart shows each slot's average response time as a bar, scaled to the slowest slot.
// On small screens only the most recent 30 slots are shown.
export function ResponseTimeChart({ slots, startLabel, endLabel }: { slots: Slot[]; startLabel: string; endLabel: string }) {
  const max = Math.max(1, ...slots.filter((s) => s.total > 0).map((s) => s.avgMs));
  const mobileFrom = slots.length - MOBILE_SLOTS;
  const dense = slots.length > 120;

  return (
    <div>
      <div className="flex gap-3">
        <div className="flex h-28 flex-col justify-between py-0.5 text-right font-mono text-[11px] text-muted">
          <span>{milliseconds(max)}</span>
          <span>0 ms</span>
        </div>
        <ul
          className={`flex h-28 flex-1 items-end ${dense ? "gap-px" : "gap-[2px]"} border-b border-l border-line pl-1`}
          aria-label={`Average response time from ${startLabel} to ${endLabel}`}
        >
          {slots.map((slot, i) => {
            const label =
              slot.total > 0
                ? `${slot.label}: average ${milliseconds(slot.avgMs)}, slowest ${milliseconds(slot.maxMs)}`
                : `${slot.label}: no data`;
            return (
              <li
                key={slot.key}
                title={label}
                aria-label={label}
                className={`flex h-full min-w-px flex-1 items-end ${i < mobileFrom ? "hidden md:flex" : ""}`}
              >
                <span
                  className={`w-full ${dense ? "rounded-t-[1px]" : "rounded-t-[2px]"} ${slot.total > 0 ? "bg-accent hover:bg-accent-strong" : ""}`}
                  style={{ height: slot.total > 0 ? `${Math.max(2, (slot.avgMs / max) * 100)}%` : 0 }}
                />
              </li>
            );
          })}
        </ul>
      </div>
      <div className="mt-1.5 flex justify-between pl-14 text-xs text-muted">
        <span>
          <span className="md:hidden">{slots.length > MOBILE_SLOTS ? `Last ${MOBILE_SLOTS} shown` : startLabel}</span>
          <span className="hidden md:inline">{startLabel}</span>
        </span>
        <span>{endLabel}</span>
      </div>
    </div>
  );
}
