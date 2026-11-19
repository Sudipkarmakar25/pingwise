# Pingwise

**Uptime monitoring for the services your team depends on.**

Pingwise checks your HTTP endpoints on a schedule, records every result, shows what is up, what is down and why, and emails the right people when a service goes down or recovers. It is multi-organisation from the ground up: many teams share one installation, and each sees only its own services, history and people.

Free and self-hosted. One Postgres database, one Go API, one Go worker, one Next.js web app.

---

## Contents

- [What it does](#what-it-does)
- [Architecture](#architecture)
- [Why the design holds up](#why-the-design-holds-up)
- [Using Pingwise in larger organisations](#using-pingwise-in-larger-organisations)
- [How far it scales](#how-far-it-scales)
- [Getting started](#getting-started)
- [Configuration](#configuration)
- [Project structure](#project-structure)
- [Roadmap](#roadmap)

---

## What it does

**Monitoring**
- Checks any HTTP endpoint with `GET`, `POST` or `HEAD`, custom headers and a request body.
- Optional auth header and token, **encrypted at rest** (AES-256-GCM, versioned keys) and never shown again after saving.
- Per-service settings: expected status code, interval (1, 2, 3, 5, 10, 15, 30 or 60 minutes), timeout (1–30 s) and failure threshold (1–5).
- A service is marked **down only after N failures in a row**, so a single network blip doesn't count as an outage.
- Every failure records its reason: wrong status code, timeout, DNS failure, refused connection, TLS error, too many redirects.

**History**
- 90-day daily uptime bars, coloured on a continuous scale from 100% to 0%.
- Detail view by minute (last 3 hours), every 5 minutes (last 24 hours) or day (last 90 days), for both uptime and response time.
- Every individual check, newest first, paged.

**Alerting**
- An email when a service **goes down** and when it **recovers**. No repeats while it stays down.
- Per-person, per-service alert settings:
  - **Owners**: on by default, can turn their own off.
  - **Editors and viewers**: off by default, can turn their own on.
  - A higher role can change a lower role's setting.
- Each email is its own job with automatic retries, sent over any SMTP server (Gmail, Brevo, SES, your company relay).

**Teams and access**
- Organisations with three roles:
  - **Owner**: manages people and services.
  - **Editor**: manages services.
  - **Viewer**: read-only.
- Owners add people with a temporary password, which must be changed on first sign-in.
- Every request is checked against the caller's organisation and role. Organisations can't see each other's data.
- Opaque access tokens stored only as SHA-256 hashes, kept in an HTTP-only cookie. Removing a member ends their sessions at once.

---

## Architecture

```mermaid
flowchart TB
    B[Browser] --> W[Next.js web]
    W --> A[Go API]
    A <--> DB[(PostgreSQL<br/>services · check_results · daily_stats<br/>alert_subscriptions · river_job queue)]

    subgraph Worker ["Worker · 1…N replicas"]
        T((":00<br/>every min")) --> S([schedule_checks])
        S -->|1 job per service| C([check_service ×N])
        C <-->|HTTP| E[Endpoint]
        C --> F{status<br/>flipped?}
        F -. no .-> T
        F -->|yes| AL([send_alert])
        AL -->|1 per recipient| EM([send_email ×N])
        EM -->|retry ×5| EM
    end

    S -->|claim due| DB
    C -->|save result| DB
    AL -->|who gets it| DB
    EM --> SMTP[SMTP] --> IN[Inbox]
```

The **API** only stores settings and serves results; it never calls a monitored URL. All monitoring happens in the **worker**, which uses [River](https://riverqueue.com), a job queue that lives in Postgres.

### The four jobs

| Job | What it does |
|---|---|
| `schedule_checks` | Wakes at :00 every minute, claims every service that is due, and creates one check job for each. |
| `check_service` | Calls one endpoint, decides up or down, and saves the result, daily rollup and new status together. Starts an alert if the status changed. |
| `send_alert` | Finds who should be told (owners plus anyone who turned alerts on) and creates one email job per person. |
| `send_email` | Sends one email to one person, retrying up to 5 times with backoff if the mail server fails. |

### How a service becomes due

The scheduler claims due services in a single statement:

```sql
UPDATE services SET last_scheduled_at = date_trunc('minute', now())
WHERE is_active AND deleted_at IS NULL
  AND (last_scheduled_at IS NULL
       OR last_scheduled_at + make_interval(mins => interval_minutes) <= date_trunc('minute', now()))
RETURNING id, org_id
```

Finding and marking happen in the same statement, so two schedulers can never book the same service. Times are rounded to the minute, so a service on a 5-minute interval runs at :00, :05, :10 and never drifts.

### How status is decided

```
pass                            → status up,   failures = 0
fail, failures + 1 ≥ threshold  → status down, failures + 1
fail, failures + 1 < threshold  → status unchanged
```

The old and new status come back from the same locked update, so each change is noticed exactly once. Alerts fire on `→ down` and `down → up` only.

---

## Why the design holds up

| Property | How |
|---|---|
| **Decoupled** | The API and the worker share only the database. Either can restart, deploy or scale without the other. |
| **No extra infrastructure** | The job queue is a Postgres table. One stateful system to run, back up and monitor. No Redis, no broker. |
| **Horizontal workers** | Run more worker replicas for more checks in parallel. River's leader election keeps the minute scheduler single, and jobs are fetched with `FOR UPDATE SKIP LOCKED`, so replicas never take the same job. |
| **No double work** | The claim is one atomic statement; check jobs are unique per service, so a slow service can't pile up queued checks. |
| **Failure isolation** | Each check and each email is its own job with its own timeout. A hanging endpoint or a failing inbox affects nothing else. |
| **Consistent data** | A check's result, the day's rollup and the service status are written in one transaction. |
| **Cheap long history** | 90-day views read one row per service per day (`service_daily_stats`), never millions of raw checks. |
| **Tenant isolation** | Every query is scoped by organisation ID taken from the authenticated session, not from the request. |

---

## Using Pingwise in larger organisations

Metrics, logs and traces tell you what a service reports about itself. Pingwise answers a different question from the outside: **can this endpoint be reached, does it answer correctly, and how fast, right now?** That black-box view catches what internal telemetry misses: expired certificates, DNS mistakes, broken load balancers, a deploy that took a route down, or a dependency that went quiet.

How the current features map onto a larger organisation:

- **One installation, many teams.** Each organisation in Pingwise can be a team, product line or business unit, each with its own services, history and people, fully separated.
- **Ownership and on-call routing.** Owners are alerted by default. Engineers opt in per service, so alerts reach the people who own that service and not everyone.
- **Monitor private and authenticated endpoints.** Self-host inside your network to check internal APIs that aren't reachable from the internet. Auth tokens for protected health endpoints are encrypted at rest with rotatable key versions.
- **Least-privilege access.** Viewers (support, management, stakeholders) see status and history without being able to change anything.
- **Fits standard operations.** Stateless API and worker processes that suit containers and Kubernetes. Postgres is the only stateful dependency, and existing backup and HA tooling applies to it as-is.
- **Third-party dependency watch.** Track the external APIs your product relies on (payments, email, identity) alongside your own, with failure reasons recorded for each outage.

Features large organisations typically need that Pingwise **does not have yet** are listed in the [Roadmap](#roadmap): SSO, Slack and webhook alerts, retention policies, multi-region checks, and metric export.

---

## How far it scales

These are **design estimates, not benchmark results**.

Each worker process runs `CHECK_WORKERS` checks at once (default 20). Checks mostly wait on the network, so:

```
checks per minute ≈ CHECK_WORKERS × 60 ÷ average check time (seconds)
```

| Setup | Checks/min (0.5 s average) | Services at 1-min interval | Services at 5-min interval |
|---|---|---|---|
| 1 worker, 20 concurrent | ~2,400 | ~2,000 | ~10,000 |
| 1 worker, 100 concurrent | ~12,000 | ~10,000 | ~50,000 |
| 3 workers, 100 concurrent each | ~36,000 | ~30,000 | ~150,000 |

What to raise first, in order:
1. **Slow or timing-out endpoints** hold a slot for up to their timeout, so raise `CHECK_WORKERS` or add worker replicas.
2. **Database connections**: raise `MAX_OPEN_CONNS_TO_DB` along with `CHECK_WORKERS`.
3. **Raw check history** grows by about 1.4 million rows a day per 1,000 one-minute services. Plan a retention policy (see the roadmap).

---

## Getting started

### Prerequisites
- Go 1.26+
- PostgreSQL 15+
- [Bun](https://bun.sh) (or Node.js 20+)
- An SMTP account for alert emails (optional; alerts are off without it)

### 1. Database
```bash
createdb pingwise
```

### 2. Backend
```bash
cd backend
cp .env.example .env
```
Edit `.env`:
- Set `PRIMARY_DB_DSN` to your database.
- Generate an encryption key with `openssl rand -base64 32` and put it in `ENCRYPTION_KEY`.
- Optionally fill in the SMTP settings.

Then create the tables:
```bash
go run ./migrate up          # Pingwise tables
go run ./cmd/river-migrate   # job queue tables
```

Run the API and the worker in separate terminals:
```bash
go run ./cmd/api             # http://localhost:8090
go run ./cmd/worker
```

### 3. Web
```bash
cd web
cp .env.example .env.local
bun install
bun run dev                  # http://localhost:3100
```

Open http://localhost:3100, create an account (you become the owner of a new organisation), and add your first service. It is checked within a minute.

### Using Pingwise

1. **Create an account.** Signing up creates your organisation and makes you its owner.
2. **Add your team.** On **Team**, add people as owner, editor or viewer with a temporary password. They choose their own password on first sign-in.
3. **Add a service.** On **Services → Add service**, enter:
   - the URL and method
   - optionally headers, a request body or an auth token
   - the expected status code, interval, timeout and failure threshold

   The first check runs within a minute.
4. **Watch status.** The dashboard lists every service with its current status, last response time and 90-day uptime bars. Click **Refresh** for the latest results.
5. **Investigate.** Open a service to see uptime and response time by minute, 5 minutes or day, plus every individual check with its failure reason.
6. **Choose who gets alerts.** Click **Notify** on a service to get emails for it. The **Alert emails** panel on the service page shows who is notified, and lets you change it for people in a lower role.
7. **Pause or edit.** On a service's edit page, change its settings, pause or resume checks, or delete it.

### Sending alert emails with Gmail
1. Turn on 2-Step Verification on the Gmail account.
2. Create an app password at https://myaccount.google.com/apppasswords.
3. In `backend/.env`, set:
   - `SMTP_HOST=smtp.gmail.com`
   - `SMTP_PORT=587`
   - `SMTP_USERNAME` to the Gmail address
   - `SMTP_PASSWORD` to the app password, without spaces
   - `ALERT_FROM_EMAIL="Pingwise <the same address>"`
4. Restart the worker.

---

## Configuration

### Backend (`backend/.env`)

| Variable | Default | Purpose |
|---|---|---|
| `ENVIRONMENT` | | `local` or `production`. When unset, settings are read from `.env` |
| `SERVER_ADDRESS` | | API listen address, e.g. `:8090` |
| `PRIMARY_DB_DSN` | | Postgres connection string |
| `MAX_OPEN_CONNS_TO_DB` | `10` | Connection pool size per process |
| `LOG_LEVEL` | | `debug`, `info`, `warn` or `error` |
| `ENCRYPTION_KEY` | | 32 random bytes, base64. **Keep it safe: losing it makes stored tokens unreadable** |
| `ENCRYPTION_KEY_VERSION` | `1` | Version stored with each encrypted token, for key rotation |
| `CHECK_WORKERS` | `20` | Checks run at the same time per worker process |
| `SMTP_HOST` | | Leave empty to turn alert emails off |
| `SMTP_PORT` | `587` | STARTTLS is used when the server offers it |
| `SMTP_USERNAME` / `SMTP_PASSWORD` | | SMTP login |
| `ALERT_FROM_EMAIL` | | Sender, e.g. `"Pingwise <alerts@example.com>"` |
| `APP_URL` | | Public address of the web app, used for links in emails |

### Web (`web/.env.local`)

| Variable | Purpose |
|---|---|
| `API_BASE_URL` | Where the Next.js server reaches the Go API, e.g. `http://localhost:8090` |

The session cookie is marked `Secure` in production builds (`bun run build && bun run start`), so serve the web app over HTTPS.

---

## Project structure

```
pingwise/
├── backend/
│   ├── cmd/
│   │   ├── api/             # HTTP API (Echo): routes, handlers, middleware
│   │   ├── worker/          # River worker process
│   │   └── river-migrate/   # creates River's queue tables
│   ├── async/
│   │   ├── args/            # job definitions (schedule_checks, check_service, send_alert, send_email)
│   │   ├── workers/         # job handlers
│   │   └── river/           # client, registry and periodic schedule
│   ├── internal/
│   │   ├── models/          # GORM models
│   │   ├── repo/            # database access
│   │   └── service/         # business logic (auth, users, monitors, checks, alerts, notifications)
│   ├── common/              # logger, database, encryption, email, request context
│   ├── config/              # environment loading for API and worker
│   └── migrate/             # dbmate runner and per-table SQL migrations
└── web/                     # Next.js app (App Router, Tailwind)
    └── src/
        ├── app/             # pages and the API proxy route handlers
        ├── components/      # UI components
        └── lib/             # API client, types, formatting
```

**Stack:** Go · Echo · GORM · River · PostgreSQL · Next.js · React · Tailwind CSS

---

## Roadmap

- [ ] **Retention policy** for raw check results (daily rollups are kept separately).
- [ ] **Outbound address protection**: block checks to private and internal addresses on shared, public installations.
- [ ] **More alert channels**: Slack, Microsoft Teams, generic webhooks.
- [ ] **Incidents**: group an outage into one record with start, end, duration and cause.
- [ ] **Public status pages** per organisation.
- [ ] **SSO** (SAML / OIDC) and an audit log.
- [ ] **More check types**: TCP, DNS, SSL certificate expiry, response body assertions.
- [ ] **Metric export** for Prometheus and OpenTelemetry.
- [ ] **Multi-region checks** with confirmation from a second location.
- [ ] **Docker Compose** and Kubernetes deployment files.
