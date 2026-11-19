import type { ReactNode } from "react";
import { ButtonLink } from "./ui/button";

// BackLink is the light "← Services" button at the top of inner pages
export function BackLink({ href, children }: { href: string; children: ReactNode }) {
  return (
    <ButtonLink href={href} variant="muted" size="sm" className="group h-7 max-w-full gap-1.5 px-2.5 text-[13px]">
      <span aria-hidden="true" className="transition-transform group-hover:-translate-x-0.5">
        ←
      </span>
      <span className="truncate">{children}</span>
    </ButtonLink>
  );
}
