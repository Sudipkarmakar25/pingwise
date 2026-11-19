import { Brand } from "./brand";
import { SessionAware } from "./session-aware";
import { ButtonLink } from "./ui/button";

const sections = [
  ["How it works", "#how-it-works"],
  ["Capabilities", "#capabilities"],
  ["Security", "#security"],
  ["FAQ", "#faq"],
];

// Header for public pages, drawn on the navy hero. Signed-in visitors don't see the sign-in and
// sign-up buttons; the hero gives them the way back to their services.
export function SiteHeader() {
  return (
    <header className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
      <Brand onDark />
      <nav aria-label="Sections" className="hidden items-center gap-7 text-sm text-hero-muted md:flex">
        {sections.map(([label, href]) => (
          <a key={href} href={href} className="transition-colors hover:text-hero-ink">
            {label}
          </a>
        ))}
      </nav>
      <SessionAware
        signedIn={null}
        signedOut={
          <div className="flex items-center gap-2">
            <ButtonLink href="/login" variant="inverse" size="sm">
              Sign in
            </ButtonLink>
            <div className="hidden sm:block">
              <ButtonLink href="/signup" variant="primary" size="sm">
                Create an account
              </ButtonLink>
            </div>
          </div>
        }
      />
    </header>
  );
}
