import type { ReactNode } from "react";

type Tone = "error" | "info" | "success";

const tones: Record<Tone, string> = {
  error: "border-down/30 bg-down-soft text-down-ink",
  info: "border-line bg-canvas text-ink",
  success: "border-up/30 bg-up-soft text-up-ink",
};

export function Alert({ tone = "info", children }: { tone?: Tone; children: ReactNode }) {
  return (
    <div role={tone === "error" ? "alert" : "status"} className={`rounded-md border px-3.5 py-3 text-sm ${tones[tone]}`}>
      {children}
    </div>
  );
}
