import { SessionAware } from "@/components/session-aware";
import { SiteHeader } from "@/components/site-header";
import { ButtonLink } from "@/components/ui/button";
import { DashboardPreview } from "./_home/dashboard-preview";
import { Icon } from "./_home/icons";

const steps = [
  ["Create your organisation", "Sign up, then invite your team as owners, editors or viewers."],
  ["Add your endpoints", "Set the URL, method, headers, credentials and the response you expect."],
  ["Review status and history", "See what is up, what is down and why, with 90 days of daily history."],
];

const capabilities = [
  {
    icon: "clock",
    title: "Scheduled checks",
    body: "Each endpoint runs on its own interval, from every minute to every hour, using GET, POST or HEAD with custom headers and a request body.",
  },
  {
    icon: "filter",
    title: "Failure thresholds",
    body: "An endpoint is marked down only after the number of consecutive failures you choose, so a single network blip doesn't read as an outage.",
  },
  {
    icon: "history",
    title: "90-day history",
    body: "Daily uptime bars, an uptime percentage and response times for every endpoint, with every individual check available to review.",
  },
  {
    icon: "search",
    title: "Failure diagnostics",
    body: "Every failed check records its reason: wrong status code, timeout, DNS failure, refused connection, TLS error or redirect loop.",
  },
  {
    icon: "lock",
    title: "Encrypted credentials",
    body: "Tokens your endpoints require are encrypted before storage and never shown again after you save them.",
  },
  {
    icon: "users",
    title: "Role-based access",
    body: "Owners manage the team, editors manage endpoints and viewers can review status. Each organisation sees only its own data.",
  },
] as const;

const security = [
  "Each organisation's services, history and team are kept separate. Nobody outside your organisation can see or change them.",
  "Every request is checked against your organisation and your role, so viewers can't make changes and only owners can manage the team.",
  "Tokens for your endpoints are encrypted when saved, used only when a check runs, and never shown again or written to logs.",
  "Passwords are never stored in readable form, and removing a team member signs them out everywhere straight away.",
];

const faqs = [
  [
    "Is Pingwise free?",
    "Yes. Pingwise is provided free of charge on a best-effort basis. It isn't a replacement for a paid monitoring service with an uptime guarantee.",
  ],
  [
    "How often are endpoints checked?",
    "Each endpoint has its own interval: every 1, 2, 3, 5, 10, 15, 30 or 60 minutes.",
  ],
  [
    "What counts as an outage?",
    "A check fails when the endpoint returns a status code other than the one you expect, doesn't respond within its timeout, or can't be reached. The endpoint is marked down after the number of consecutive failures you set, from 1 to 5, and marked up again as soon as a check passes.",
  ],
  [
    "How are my credentials handled?",
    "They're encrypted before they're stored, decrypted only when a check runs, and never displayed or logged. Where you can, use a dedicated read-only token for monitoring.",
  ],
  [
    "Can other organisations see my services?",
    "No. Each organisation's services, history and team are kept separate, and every request is checked against the organisation it belongs to. Only people an owner adds can sign in to your organisation.",
  ],
  [
    "Where are checks run from?",
    "From a single location. A network problem between Pingwise and your endpoint can therefore appear as a failed check.",
  ],
  ["Does Pingwise send alerts?", "Not at the moment. Status and history are available in the dashboard."],
];

export default function HomePage() {
  return (
    <>
      <div className="bg-hero">
        <SiteHeader />
        <section className="mx-auto grid max-w-6xl items-center gap-12 px-6 pt-12 pb-20 lg:grid-cols-[1.05fr_1fr] lg:pt-20 lg:pb-28">
          <div>
            <p className="text-sm font-medium text-accent">Uptime monitoring</p>
            <h1 className="mt-3 text-4xl font-semibold tracking-tight text-hero-ink sm:text-5xl sm:leading-[1.1]">
              Uptime monitoring for the services your team depends on
            </h1>
            <p className="mt-5 max-w-xl text-lg leading-relaxed text-hero-muted">
              Pingwise checks your endpoints on a schedule, records every result, and shows your team what is up,
              what is down, and why.
            </p>
            <div className="mt-8">
              <SessionAware
                signedIn={
                  <div className="flex flex-wrap gap-3">
                    <ButtonLink href="/dashboard" variant="primary" size="lg">
                      Go to your services
                      <span aria-hidden="true">→</span>
                    </ButtonLink>
                    <ButtonLink href="/services/new" variant="inverse" size="lg">
                      Add a service
                    </ButtonLink>
                  </div>
                }
                signedOut={
                  <div className="flex flex-wrap gap-3">
                    <ButtonLink href="/signup" variant="primary" size="lg">
                      Create an account
                    </ButtonLink>
                    <ButtonLink href="/login" variant="inverse" size="lg">
                      Sign in
                    </ButtonLink>
                  </div>
                }
              />
            </div>
          </div>
          <DashboardPreview />
        </section>
      </div>

      <main>
        <section id="how-it-works" className="mx-auto max-w-6xl scroll-mt-8 px-6 py-20">
          <SectionHeading eyebrow="How it works" title="Set up in a few minutes" />
          <ol className="mt-10 grid gap-6 md:grid-cols-3">
            {steps.map(([title, body], i) => (
              <li key={title} className="rounded-xl border border-line bg-surface p-6">
                <span className="flex h-8 w-8 items-center justify-center rounded-full bg-hero font-mono text-sm text-hero-ink">
                  {i + 1}
                </span>
                <h3 className="mt-4 font-semibold">{title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-muted">{body}</p>
              </li>
            ))}
          </ol>
        </section>

        <section id="capabilities" className="scroll-mt-8 border-y border-line bg-surface">
          <div className="mx-auto max-w-6xl px-6 py-20">
            <SectionHeading
              eyebrow="Capabilities"
              title="What Pingwise does"
              lead="A focused set of tools for knowing whether your services are available, and why they aren't."
            />
            <div className="mt-10 grid gap-x-10 gap-y-9 sm:grid-cols-2 lg:grid-cols-3">
              {capabilities.map((c) => (
                <div key={c.title}>
                  <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-canvas text-accent-strong ring-1 ring-line">
                    <Icon name={c.icon} />
                  </span>
                  <h3 className="mt-4 font-semibold">{c.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-muted">{c.body}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="security" className="mx-auto grid max-w-6xl scroll-mt-8 gap-10 px-6 py-20 lg:grid-cols-[1fr_1.4fr]">
          <SectionHeading
            eyebrow="Security"
            title="Your data stays with your organisation"
            lead="Many teams use Pingwise, but each organisation only ever sees its own services, history and people."
          />
          <ul className="space-y-4">
            {security.map((item) => (
              <li key={item} className="flex gap-3 rounded-xl border border-line bg-surface p-4 text-sm leading-relaxed">
                <Icon name="check" className="mt-0.5 h-5 w-5 shrink-0 text-up" />
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </section>

        <section id="faq" className="scroll-mt-8 border-t border-line bg-surface">
          <div className="mx-auto max-w-3xl px-6 py-20">
            <SectionHeading eyebrow="FAQ" title="Common questions" centered />
            <div className="mt-10 divide-y divide-line border-y border-line">
              {faqs.map(([q, a]) => (
                <details key={q} className="group py-5">
                  <summary className="flex cursor-pointer list-none items-center justify-between gap-4 font-medium">
                    {q}
                    <span className="text-muted transition-transform group-open:rotate-45" aria-hidden="true">
                      +
                    </span>
                  </summary>
                  <p className="mt-3 text-sm leading-relaxed text-muted">{a}</p>
                </details>
              ))}
            </div>
          </div>
        </section>

        <section className="bg-hero">
          <div className="mx-auto flex max-w-6xl flex-col items-start justify-between gap-6 px-6 py-16 md:flex-row md:items-center">
            <SessionAware
              signedIn={
                <div>
                  <h2 className="text-2xl font-semibold tracking-tight text-hero-ink">Your services are waiting</h2>
                  <p className="mt-2 text-hero-muted">See their status, history and recent checks.</p>
                </div>
              }
              signedOut={
                <div>
                  <h2 className="text-2xl font-semibold tracking-tight text-hero-ink">Start monitoring your services</h2>
                  <p className="mt-2 text-hero-muted">Create an organisation and add your first endpoint.</p>
                </div>
              }
            />
            <SessionAware
              signedIn={
                <ButtonLink href="/dashboard" variant="primary" size="lg">
                  Go to your services
                  <span aria-hidden="true">→</span>
                </ButtonLink>
              }
              signedOut={
                <ButtonLink href="/signup" variant="primary" size="lg">
                  Create an account
                </ButtonLink>
              }
            />
          </div>
        </section>
      </main>
    </>
  );
}

function SectionHeading({
  eyebrow,
  title,
  lead,
  centered = false,
}: {
  eyebrow: string;
  title: string;
  lead?: string;
  centered?: boolean;
}) {
  return (
    <div className={centered ? "text-center" : ""}>
      <p className="text-sm font-medium text-accent-strong">{eyebrow}</p>
      <h2 className="mt-2 text-3xl font-semibold tracking-tight">{title}</h2>
      {lead && <p className={`mt-3 max-w-xl leading-relaxed text-muted ${centered ? "mx-auto" : ""}`}>{lead}</p>}
    </div>
  );
}
