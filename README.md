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
- [Getting started](#getting-started): [on your computer](#run-it-on-your-computer) or [self-host on a server](#self-host-on-a-server)
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

There are two ways to run Pingwise:
- **[On your computer](#run-it-on-your-computer)**, to try it out.
- **[On a server](#self-host-on-a-server)**, for real use: always on, HTTPS and your own domain.

### Run it on your computer

**You need:** Go 1.26+, PostgreSQL 14+, [Bun](https://bun.sh) and Node.js 20+.

```bash
git clone https://github.com/Sudipkarmakar25/pingwise.git
cd pingwise
createdb pingwise
```

**Backend:**
```bash
cd backend
cp .env.example .env
```
Edit `.env`:
- Set `PRIMARY_DB_DSN` to your database.
- Put the output of `openssl rand -base64 32` in `ENCRYPTION_KEY`.

Then create the tables:
```bash
go run ./migrate up          # Pingwise tables
go run ./cmd/river-migrate   # job queue tables
```

Run the API and the worker in two terminals:
```bash
go run ./cmd/api             # http://localhost:8090
go run ./cmd/worker
```

**Web**, in a third terminal:
```bash
cd web
cp .env.example .env.local
bun install
bun run dev                  # http://localhost:3100
```

Open http://localhost:3100 and create an account.

---

### Self-host on a server

This guide sets Pingwise up on one **Ubuntu 24.04** server with HTTPS:
- Postgres, the API, the worker and the web app all run on the same machine.
- [Caddy](https://caddyserver.com) serves your domain and gets the HTTPS certificate automatically.
- systemd keeps everything running and restarts it after a crash or reboot.

Only ports 80 and 443 are open to the internet. The API and the database are reachable only from the server itself.

**You need:**
- **A server** with 2 GB RAM. 1 GB works if you add swap (step 1).
- **A domain or subdomain**, e.g. `status.example.com`, with a DNS **A record** pointing at the server's IP address.
- **SSH access** with `sudo`.

In the commands below, replace `status.example.com` with your domain and `CHANGE_ME_DB_PASSWORD` with a strong password.

#### 1. Install the software
```bash
sudo apt update
sudo apt install -y git unzip postgresql caddy

# Node.js 22, used to run the web app
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt install -y nodejs

# Go 1.26. Use linux-arm64 instead of linux-amd64 on ARM servers (e.g. Oracle Ampere)
curl -fsSL https://go.dev/dl/go1.26.4.linux-amd64.tar.gz | sudo tar -C /usr/local -xz
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
```

On a 1 GB server, add swap so the web build doesn't run out of memory:
```bash
sudo fallocate -l 2G /swapfile && sudo chmod 600 /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

#### 2. Create the database
```bash
sudo -u postgres psql -c "CREATE USER pingwise WITH PASSWORD 'CHANGE_ME_DB_PASSWORD';"
sudo -u postgres psql -c "CREATE DATABASE pingwise OWNER pingwise;"
```

#### 3. Create a user for Pingwise and download the code
Pingwise runs as its own user, not as root.
```bash
sudo useradd --create-home --shell /bin/bash pingwise
sudo -iu pingwise                                   # you are now the pingwise user
git clone https://github.com/Sudipkarmakar25/pingwise.git
curl -fsSL https://bun.sh/install | bash            # installs Bun for this user
```

#### 4. Configure the backend
Still as the `pingwise` user:
```bash
cd ~/pingwise/backend
cp .env.example .env
chmod 600 .env
openssl rand -base64 32                             # copy this output
nano .env
```

Set these values in `.env` and leave the rest as they are:
```ini
ENVIRONMENT=production
SERVER_ADDRESS=127.0.0.1:8090
PRIMARY_DB_DSN=postgres://pingwise:CHANGE_ME_DB_PASSWORD@localhost:5432/pingwise?sslmode=disable
ENCRYPTION_KEY=<the openssl output>
APP_URL=https://status.example.com
```

For alert emails, also fill in the `SMTP_…` and `ALERT_FROM_EMAIL` lines (see [Sending alert emails with Gmail](#sending-alert-emails-with-gmail)). Without them, Pingwise works but sends no emails.

> **Keep a copy of `ENCRYPTION_KEY` somewhere safe.** Without it, saved service tokens can't be decrypted.

#### 5. Build the backend and create the tables
```bash
cd ~/pingwise/backend
go build -o bin/api ./cmd/api
go build -o bin/worker ./cmd/worker
go build -o bin/migrate ./migrate
go build -o bin/river-migrate ./cmd/river-migrate
./bin/migrate up
./bin/river-migrate
```

#### 6. Build the web app
```bash
cd ~/pingwise/web
echo 'API_BASE_URL=http://127.0.0.1:8090' > .env.local
~/.bun/bin/bun install
~/.bun/bin/bun run build
exit                                                # back to your own user
```

#### 7. Run everything as services
Create three systemd services, so Pingwise starts on boot and restarts if it stops:
```bash
sudo tee /etc/systemd/system/pingwise-api.service > /dev/null <<'EOF'
[Unit]
Description=Pingwise API
After=network-online.target postgresql.service
Wants=network-online.target

[Service]
User=pingwise
WorkingDirectory=/home/pingwise/pingwise/backend
EnvironmentFile=/home/pingwise/pingwise/backend/.env
ExecStart=/home/pingwise/pingwise/backend/bin/api
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo tee /etc/systemd/system/pingwise-worker.service > /dev/null <<'EOF'
[Unit]
Description=Pingwise worker
After=network-online.target postgresql.service
Wants=network-online.target

[Service]
User=pingwise
WorkingDirectory=/home/pingwise/pingwise/backend
EnvironmentFile=/home/pingwise/pingwise/backend/.env
ExecStart=/home/pingwise/pingwise/backend/bin/worker
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo tee /etc/systemd/system/pingwise-web.service > /dev/null <<'EOF'
[Unit]
Description=Pingwise web
After=network-online.target pingwise-api.service
Wants=network-online.target

[Service]
User=pingwise
WorkingDirectory=/home/pingwise/pingwise/web
Environment=NODE_ENV=production
ExecStart=/home/pingwise/.bun/bin/bun run start
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now pingwise-api pingwise-worker pingwise-web
systemctl status pingwise-api pingwise-worker pingwise-web --no-pager
```
All three should show **active (running)**.

#### 8. Turn on HTTPS
Point Caddy at the web app. It gets and renews the certificate on its own.
```bash
sudo tee /etc/caddy/Caddyfile > /dev/null <<'EOF'
status.example.com {
    reverse_proxy 127.0.0.1:3100
}
EOF
sudo systemctl reload caddy
```

#### 9. Open only the ports you need
```bash
sudo ufw allow OpenSSH
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable
```
On a cloud server (Oracle, AWS, GCP, Azure), also allow ports 80 and 443 in the provider's firewall (security list or security group).

#### 10. Sign up
Open `https://status.example.com` and **create an account**. You become the owner of the first organisation. Then add your team and your services.

> **Who can sign up:** anyone who can open your site can create their own organisation. On a public address, put the site behind your VPN or company network if only your team should use it. Pingwise doesn't yet block checks to private network addresses (see the roadmap), which matters when strangers can add services.

#### Logs
```bash
journalctl -u pingwise-worker -f       # checks and emails
journalctl -u pingwise-api -f          # API requests and errors
journalctl -u pingwise-web -f          # web app
```

#### Updating to a new version
```bash
sudo -iu pingwise
cd ~/pingwise && git pull

cd backend
go build -o bin/api ./cmd/api && go build -o bin/worker ./cmd/worker
go build -o bin/migrate ./migrate && go build -o bin/river-migrate ./cmd/river-migrate
./bin/migrate up && ./bin/river-migrate

cd ../web
~/.bun/bin/bun install && ~/.bun/bin/bun run build
exit

sudo systemctl restart pingwise-api pingwise-worker pingwise-web
```

#### Backups
Back up the database daily and keep copies off the server:
```bash
sudo -u postgres pg_dump -Fc pingwise > pingwise-$(date +%F).dump
```
To restore into an empty `pingwise` database:
```bash
sudo -u postgres pg_restore -d pingwise pingwise-YYYY-MM-DD.dump
```
Also keep a copy of `backend/.env`, or at least `ENCRYPTION_KEY`, in a password manager.

---

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
Any SMTP server works. These steps are for Gmail.

1. Turn on 2-Step Verification on the Gmail account.
2. Create an app password at https://myaccount.google.com/apppasswords.
3. In `backend/.env`, set:
   - `SMTP_HOST=smtp.gmail.com`
   - `SMTP_PORT=587`
   - `SMTP_USERNAME` to the Gmail address
   - `SMTP_PASSWORD` to the app password, without spaces
   - `ALERT_FROM_EMAIL="Pingwise <the same address>"`
4. Restart the worker: `sudo systemctl restart pingwise-worker` on a server, or stop and start `go run ./cmd/worker` locally.

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
