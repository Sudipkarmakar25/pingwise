import Link from "next/link";

export function LogoMark({ className = "h-7 w-7" }: { className?: string }) {
  return (
    <svg viewBox="0 0 32 32" className={className} aria-hidden="true">
      <rect width="32" height="32" rx="7" fill="#0F2236" />
      <path
        d="M5 17h6l3-7 4 13 3-6h6"
        fill="none"
        stroke="#14B8A6"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

// Brand is the logo and name, linking home. onDark switches the text to the hero colours.
export function Brand({ onDark = false }: { onDark?: boolean }) {
  return (
    <Link href="/" className="flex items-center gap-2.5" aria-label="Pingwise home">
      <LogoMark className={`h-7 w-7 ${onDark ? "rounded-md ring-1 ring-white/15" : ""}`} />
      <span className={`text-[17px] font-semibold tracking-tight ${onDark ? "text-hero-ink" : "text-ink"}`}>
        Pingwise
      </span>
    </Link>
  );
}
