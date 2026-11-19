// Colour for a day's uptime on a continuous scale: dark green at 100%, through light green,
// yellow and red, to dark red at 0%. Most real days fall between 95% and 100%, so the stops
// are packed into that range; spacing them evenly would make a 90% day look almost green.
export const UPTIME_STOPS: [percent: number, colour: string][] = [
  [100, "#15803d"], // dark green
  [99.5, "#22c55e"], // green
  [99, "#86efac"], // light green
  [97, "#facc15"], // yellow
  [95, "#f97316"], // orange
  [90, "#ef4444"], // red
  [0, "#7f1d1d"], // dark red
];

function toRgb(hex: string) {
  const n = parseInt(hex.slice(1), 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

// uptimeColour blends between the two stops on either side of the given percentage
export function uptimeColour(percent: number) {
  const p = Math.min(100, Math.max(0, percent));
  for (let i = 0; i < UPTIME_STOPS.length - 1; i++) {
    const [highP, highC] = UPTIME_STOPS[i];
    const [lowP, lowC] = UPTIME_STOPS[i + 1];
    if (p <= highP && p >= lowP) {
      const t = highP === lowP ? 0 : (highP - p) / (highP - lowP);
      const a = toRgb(highC);
      const b = toRgb(lowC);
      const mix = a.map((v, j) => Math.round(v + (b[j] - v) * t));
      return `rgb(${mix.join(", ")})`;
    }
  }
  return UPTIME_STOPS[UPTIME_STOPS.length - 1][1];
}
