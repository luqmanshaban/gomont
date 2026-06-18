# Gomont


An open-source uptime monitor written in Go. Add the URLs you care about, get notified by email the moment one goes down, and watch status update live on a clean, dependency-light dashboard.

Gomont is built with a deliberate bias toward simplicity: a single Go binary, no JavaScript framework, no build step for the frontend, and a minimal set of external dependencies. The goal is that anyone comfortable with Go and vanilla HTML/CSS/JS can read the whole codebase in an afternoon.

<img width="1140" height="781" alt="image" src="https://github.com/user-attachments/assets/cc6adcb6-f002-4408-bd07-00c001d92d0b" /> 

## Features

- Passwordless authentication — sign up and log in with an emailed one-time code, no passwords to manage
- Add any HTTP/HTTPS endpoint and set how often it gets checked
- Configurable notification emails per account, independent of your login email
- Manual retry, editable check intervals, and one-click delete per monitor
- Live dashboard updates via Server-Sent Events — status changes appear without a page refresh
- A background worker pool that checks endpoints on schedule, retries failing ones with backoff, and emails you once a monitor is confirmed down

## Tech stack

- **Backend:** Go, `net/http` (Go 1.22+ method-and-path routing, no router library), PostgreSQL
- **Frontend:** Server-rendered HTML via `html/template`, vanilla JavaScript (ES modules, no bundler), plain CSS
- **Email:** `gopkg.in/mail.v2` over SMTP, HTML + plaintext templates
- **Live updates:** Server-Sent Events, no WebSocket dependency

Static assets and templates are embedded into the compiled binary with `//go:embed`, so deploying Gomont is copying one binary plus your `.env` file — nothing else to ship.

## Getting started

You can run Gomont with Docker Compose (recommended — handles Postgres for you) or directly with a local Go toolchain and your own Postgres instance.

### Option A: Docker Compose

**Requirements:** Docker and Docker Compose.

1. Clone the repo and copy the example environment file:

   ```bash
   git clone https://github.com/luqmanshaban/gomont.git
   cd gomont
   cp .env.example .env
   ```

2. Fill in `.env` (see [Configuration](#configuration) below). For the Docker path, `DB_HOST` in your `.env` is overridden automatically to point at the `db` service — you don't need to change it.

3. Start everything:

   ```bash
   docker compose up --build
   ```

   On first run, Postgres will automatically load `schema.sql`. The app will be available at `http://localhost:<PORT>` once both containers report healthy.

### Option B: Local Go + your own Postgres

**Requirements:** Go 1.26 or later, a running PostgreSQL instance.

1. Clone the repo and copy the example environment file:

   ```bash
   git clone https://github.com/luqmanshaban/gomont.git
   cd gomont
   cp .env.example .env
   ```

2. Create the database and load the schema:

   ```bash
   createdb gomont
   psql -d gomont -f schema.sql
   ```

3. Fill in `.env` with your local Postgres credentials (`DB_HOST` should be `localhost` here).

4. Run it:

   ```bash
   go run ./main.go
   ```

## Configuration

All configuration is via environment variables, loaded from `.env`. See `.env.example` for a template.

| Variable | Description |
|---|---|
| `DB_HOST` | Postgres host. `localhost` for local Go, `db` for Docker Compose (set automatically in compose). |
| `DB_PORT` | Postgres port, typically `5432`. |
| `DB_NAME` | Database name. |
| `DB_USER` | Database user. |
| `DB_PASS` | Database password. |
| `PORT` | Port the Gomont HTTP server listens on. |
| `JWT_SECRET` | Secret key used to sign and verify auth tokens. Use a long, random value. |
| `APP_URL` | Public base URL of your deployment (used to build links in notification emails), e.g. `http://localhost:`. |
| `EMAIL_HOST` | SMTP host for sending OTP and notification emails. |
| `EMAIL_PORT` | SMTP port. |
| `EMAIL_USER` | SMTP account username (also used as the "From" address). |
| `EMAIL_PASS` | SMTP account password or app-specific password. |

**Note on email:** if you're using Gmail, you'll need an [App Password](https://support.google.com/accounts/answer/185833) rather than your regular account password, since Gmail blocks plain password SMTP auth by default.

**Note on this version:** Gomont currently runs as a single-user deployment — each instance serves one account. There's a JWT-based auth layer in place (rather than something simpler like a static API key) specifically so the project can grow into multi-user support later without a structural rewrite.

## Project structure

```
gomont/
├── main.go                 # wiring: db, stores, worker pool, HTTP server
├── schema.sql               # Postgres schema
├── internals/
│   ├── api/                 # HTTP server, route registration, handlers
│   │   ├── handlers/         # one handler struct per resource (users, urls, notifications, pages, events)
│   │   ├── utils/             # JWT signing/verification, JSON response helpers
│   │   └── web/                # embedded frontend: templates/ and static/
│   ├── config/               # environment variable loading
│   ├── core/                  # shared domain types (e.g. the URL/monitor struct)
│   ├── store/                  # Postgres queries, one store per resource
│   ├── worker/                  # background producer/consumer pool that runs health checks
│   └── utils/                    # email sending and templates
```

## API overview

The full HTTP API is documented for contributors building alternative clients or integrations. All endpoints are JSON in and out except `GET /health`. Endpoints under auth require a `Bearer` JWT in the `Authorization` header.

| Method & Path | Description |
|---|---|
| `POST /auth` | Start signup — sends a one-time code to the given email. |
| `POST /auth/login` | Start login (send code) or complete login/signup (verify code, returns JWT), depending on payload. |
| `GET /users` | Get the authenticated user's profile. |
| `PATCH /users` | Update display name. |
| `DELETE /users` | Delete account and all associated monitors. |
| `GET /notifications/channels` | Get the authenticated user's notification email list. |
| `GET /notifications/channels/{row_id}` | Get a specific notification channel by ID. |
| `POST /notifications/channels/{row_id}` | Add one or more emails to a channel. |
| `PATCH /notifications/channels/{row_id}` | Replace one email with another. |
| `DELETE /notifications/channels/{row_id}` | Remove an email from a channel. |
| `POST /urls` | Add a new monitor. |
| `GET /urls` | List all monitors for the authenticated user. |
| `GET /urls/{url_id}` | Get a single monitor's details. |
| `PATCH /urls/{url_id}` | Update a monitor's endpoint and/or check interval. |
| `DELETE /urls/{url_id}` | Delete a monitor. |
| `POST /urls/{url_id}/retry` | Force an immediate check, bypassing the schedule. |
| `GET /events?token=<jwt>` | Server-Sent Events stream of live status changes for the dashboard. |
| `GET /health` | Database connectivity check (plain text response). |

## How it works

A **producer** goroutine polls the database every 500ms for monitors due for a check and pushes them onto a job channel. A pool of **worker** goroutines consume that channel, ping each endpoint, and update its status. If a check fails, the worker retries with backoff up to a configurable maximum before marking the monitor down and sending a notification email — once per outage, not on every failed retry.

When a monitor's status actually flips (down → healthy or healthy → down), the worker publishes an event to an in-process broker. Any dashboard with an open `/events` connection receives that update immediately and patches the relevant row in place — no polling, no full page refresh.

## Contributing

Issues and pull requests are welcome. A few things that'll make a PR easier to review:

- Keep the frontend dependency-free — no new npm packages, bundlers, or frontend frameworks. The project's whole point is staying simple enough to embed and ship as one binary.
- If you're touching `internals/api/web/`, match the existing visual language (see `static/css/base.css` for the design tokens — colors, spacing, button styles) rather than introducing a new style.
- Run `go vet ./...` before submitting.
- Open an issue first for anything that changes the database schema or the JSON shape of an existing endpoint, since those are breaking changes for anyone already running Gomont.

## License

MIT — see [LICENSE](./LICENSE) for details.
