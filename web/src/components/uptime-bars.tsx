import { percent } from "@/lib/format";
import type { Slot } from "@/lib/slots";
import { uptimeColour } from "@/lib/uptime-colour";

const MOBILE_SLOTS = 30;

function slotText(slot: Slot) {
  if (slot.total === 0) return "No data";
  const uptime = Math.round((slot.up / slot.total) * 10000) / 100;
  return `${percent(uptime)} uptime, ${slot.up.toLocaleString("en-GB")} of ${slot.total.toLocaleString("en-GB")} checks passed`;
}

// UptimeBars draws one bar per slot, coloured by the share of its checks that passed; slots without
// checks are grey. On small screens only the most recent 30 slots are shown. showLabels=false
// leaves out the start and end labels, for callers that label the bars themselves.
export function UptimeBars({
  slots,
  startLabel,
  endLabel,
  height = "h-7",
  showLabels = true,
}: {
  slots: Slot[];
  startLabel: string;
  endLabel: string;
  height?: string;
  showLabels?: boolean;
}) {
  const mobileFrom = slots.length - MOBILE_SLOTS;
  const dense = slots.length > 120;

  return (
    <div>
      <ul className={`flex ${dense ? "gap-px" : "gap-[2px]"} ${height}`} aria-label={`Uptime from ${startLabel} to ${endLabel}`}>
        {slots.map((slot, i) => {
          const label = `${slot.label}: ${slotText(slot)}`;
          const colour = slot.total > 0 ? uptimeColour((slot.up / slot.total) * 100) : undefined;
          return (
            <li
              key={slot.key}
              title={label}
              aria-label={label}
              style={colour ? { backgroundColor: colour } : undefined}
              className={`min-w-px flex-1 ${dense ? "rounded-[1px]" : "rounded-[2px]"} ${colour ? "" : "bg-nodata"} ${i < mobileFrom ? "hidden md:block" : ""} hover:opacity-70`}
            />
          );
        })}
      </ul>
      {showLabels && (
        <div className="mt-1.5 flex justify-between text-xs text-muted">
          <span>
            <span className="md:hidden">{slots.length > MOBILE_SLOTS ? `Last ${MOBILE_SLOTS} shown` : startLabel}</span>
            <span className="hidden md:inline">{startLabel}</span>
          </span>
          <span>{endLabel}</span>
        </div>
      )}
    </div>
  );
}
