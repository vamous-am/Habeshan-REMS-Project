# Habeshan REMS

**Remote Employee Management System for Ethiopian Outsourcing Agencies**

Habeshan REMS is an offline-first, multi-tenant web platform that gives Ethiopian outsourcing agencies and distributed teams verifiable visibility into attendance, task progress, and hours worked — replacing spreadsheets, chat threads, and informal check-ins with a system managers can actually trust.

Built by a team of 4 developers as an academic capstone project at Addis Ababa University (AAiT), targeting a Release 1.0 (MVP) showcase.

---

## Table of Contents

- [Why This Exists](#why-this-exists)
- [Core Features (Release 1.0 Scope)](#core-features-release-10-scope)
- [Out of Scope for MVP](#out-of-scope-for-mvp)
- [Tech Stack](#tech-stack)
- [System Architecture](#system-architecture)
- [Database Schema](#database-schema)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Project Structure](#project-structure)
- [API Overview](#api-overview)
- [Team & Ownership](#team--ownership)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Offline-First Design Notes](#offline-first-design-notes)
- [Roadmap](#roadmap)
- [Documentation](#documentation)
- [License](#license)

---

## Why This Exists

Internet connectivity in Addis Ababa and other Ethiopian markets is frequently interrupted. Most remote-work tooling assumes a stable connection, which makes it a poor fit for agencies whose employees may lose connectivity mid-shift. Habeshan REMS is built offline-first from the ground up: attendance capture *must* succeed locally regardless of network state, then reconcile with the server automatically once connectivity returns — with tamper-evident verification so a lost connection never becomes a loophole for falsified hours.

## Core Features (Release 1.0 Scope)

- **Authentication & RBAC** — email/password login, JWT-based sessions, three roles (Admin, Manager, Employee), password reset
- **Multi-organization support** — each subscribing agency operates in an isolated workspace; no cross-org data leakage at the API layer
- **Offline-capable attendance** — clock in/out with no connectivity, tamper-evident device-hash queuing, automatic sync on reconnect
- **Task management** — multi-employee task assignment, Kanban-style status tracking, per-task work timer with pause/resume and reason capture
- **Automated timesheets** — system-generated from attendance + task-timer data, with employee review/submission and manager approval workflow
- **Manager dashboard** — team attendance, task progress, pending approvals at a glance
- **Performance leaderboard** — lightweight, opt-out, named top-3 performer view
- **Telegram notifications** — reminders, approvals, and rejections pushed via Telegram Bot
- **Reporting & export** — attendance and task reports exportable to CSV, Excel, and PDF
- **ETB-denominated org & billing display** — full payment processing deferred to Release 1.1

## Out of Scope for MVP

Explicitly deferred to protect the delivery timeline (see SRS Section 2.6 for full detail):

- Automated payroll calculation and disbursement
- In-app payment processing (Chapa / Telebirr) — Release 1.1
- Screenshot capture or continuous activity surveillance
- AI-based productivity or sentiment analysis
- Git/IDE integration for automatic proof-of-work
- Native iOS/Android apps — MVP ships as an installable PWA only
- Real-time in-app chat (Telegram is the supported channel)
- Email/SMS notification channels
- Multi-level management hierarchies beyond Admin–Manager–Employee

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 18 (Vite), TypeScript, Tailwind CSS, Shadcn UI |
| Offline storage | Dexie.js over IndexedDB, AES-256 encryption at rest |
| Backend | Go 1.22+, Fiber v2 web framework, GORM v2 |
| Database | PostgreSQL 15 (Aiven Cloud, managed) |
| Auth | JWT via `golang-jwt/jwt` v5, bcrypt (cost ≥ 10) |
| Notifications | Telegram Bot API via `gopkg.in/telebot.v3` |
| Hosting | Frontend: Vercel/Netlify · Backend: Render/Railway · DB: Aiven |
| CI | GitHub Actions (lint + test on every PR) |

**Why this stack, briefly:** the frontend must be a client-rendered SPA (not SSR) because it needs to render fully offline. TypeScript gives compile-time-checked contracts between the offline queue, API client, and UI. Go/Fiber's goroutine model comfortably handles concurrent offline-batch sync and scheduled jobs on minimal server resources. Full rationale for every constraint is documented in SRS Section 2.5.

## System Architecture

Habeshan REMS is a three-tier, offline-first architecture:

```
┌─────────────────────────────┐
│   Client Tier (PWA)         │
│   React SPA + Dexie.js      │
│   offline queue (IndexedDB) │
└──────────┬───────────────────┘
           │ HTTPS / REST (JSON, Bearer JWT)
┌──────────▼───────────────────┐
│   Application Tier           │
│   Go + Fiber REST API        │
│   JWT auth middleware        │──── trigger event ───► Telegram Bot Service
└──────────┬───────────────────┘
           │ SQL (GORM)
┌──────────▼───────────────────┐
│   Data Tier                  │
│   PostgreSQL 15 (Aiven)      │
└───────────────────────────────┘
```

Clock-in/clock-out actions are captured locally first, queued in an encrypted IndexedDB store if offline, and synced automatically once connectivity returns. The server independently re-verifies every synced record's device hash and timestamp before accepting it — the client is never trusted as the source of truth for time.

## Database Schema

11 normalized tables (3NF), every organization-owned table scoped by `org_id` and enforced at the application layer:

| Table | Purpose |
|---|---|
| `organizations` | Tenant/billing root — one row per subscribing agency |
| `users` | All Admin, Manager, and Employee accounts |
| `teams` | Groups employees under a supervising Manager (Admins have global override) |
| `team_members` | Many-to-many: employees ↔ teams |
| `attendance_logs` | Clock-in/out sessions, offline-sync state, tamper-evident device hash |
| `tasks` | Units of work created by Manager/Admin |
| `task_assignments` | Many-to-many: tasks ↔ assigned employees |
| `task_time_logs` | Per-task timer events; carries the same device-hash + sync-state protection as attendance |
| `timesheets` | Auto-generated, period-based rollups with approval workflow |
| `notifications` | In-app notification log, persisted independent of Telegram delivery |
| `telegram_subscribers` | Links a user account to a Telegram chat |

Full column-level definitions, constraints, and the ER diagram live in the project SRS (Section 3.4).

> **Note on `users.email`:** uniqueness is scoped per organization, not global — `UNIQUE (org_id, email) WHERE deleted_at IS NULL`. This allows the same email to hold separate accounts under different organizations, and frees up for reuse once an account is soft-deleted.

### Core state machines

Three formally defined state machines govern the system (full transition tables in SRS Section 3.5):

1. **Attendance sync** — `OFFLINE_LOGGED → PENDING_SYNC → SYNCED_VERIFIED` or `REJECTED_TAMPERED`
2. **Task status** — `To Do → In Progress ⇄ Paused / Blocked → Completed`
3. **Timesheet approval** — `Draft → Submitted → Approved` or `Rejected → Submitted` (resubmission loop)

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 18+ and npm
- A PostgreSQL 15 instance (local or Aiven)
- A Telegram bot token (create one via [@BotFather](https://t.me/BotFather))

### Backend setup

```bash
cd backend
cp .env.example .env        # fill in your local values — see Environment Variables below
go mod download
go run ./cmd/migrate         # run GORM migrations
go run ./cmd/server           # starts the API on the configured PORT
```

### Frontend setup

```bash
cd frontend
npm install
cp .env.example .env.local  # set VITE_API_BASE_URL to your local backend
npm run dev
```

The app will be available at `http://localhost:5173` (or whichever port Vite assigns), talking to the backend at `http://localhost:8080/api/v1` by default.

### Verifying your setup

Once both are running, you should be able to register an account, log in, and see a role-appropriate (empty) dashboard shell. If you can't get this far, check the Environment Variables section below before opening an issue.

## Environment Variables

Backend `.env` (never commit real values — see `.env.example`):

| Variable | Example | Purpose |
|---|---|---|
| `PORT` | `8080` | Backend server port |
| `APP_ENV` | `development` \| `production` | Environment mode |
| `DB_HOST` | `pg-host.aivencloud.com` | Database host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `avnadmin` | Database username |
| `DB_PASSWORD` | — | Database password |
| `DB_NAME` | `habeshan_rems` | Database name |
| `JWT_SECRET` | 32+ char random string | JWT signing secret |
| `JWT_EXPIRES_IN` | `24h` | Token expiry duration |
| `FRONTEND_URL` | `https://app.habeshanrems.com` | CORS allowed origin |
| `TELEGRAM_BOT_TOKEN` | — | Telegram Bot API token |
| `RATE_LIMIT_WINDOW` | `900000` | Rate-limit window (ms), 15 min |
| `RATE_LIMIT_MAX` | `20` | Max auth requests per window per IP |

Frontend `.env.local`:

| Variable | Example | Purpose |
|---|---|---|
| `VITE_API_BASE_URL` | `http://localhost:8080/api/v1` | Backend API base URL |

## Project Structure

```
habeshan-rems/
├── backend/
│   ├── cmd/
│   │   ├── server/          # API entrypoint
│   │   └── migrate/         # migration runner
│   ├── internal/
│   │   ├── handlers/        # HTTP layer
│   │   ├── services/        # business logic + state machines
│   │   ├── repositories/    # GORM data access
│   │   ├── middleware/      # auth, RBAC, rate limiting
│   │   └── models/          # GORM structs
│   └── migrations/          # versioned schema migrations
├── frontend/
│   ├── src/
│   │   ├── features/        # feature-first structure (auth, attendance, tasks, timesheets...)
│   │   ├── components/      # shared UI components
│   │   ├── lib/              # API client, Dexie.js queue setup
│   │   └── routes/           # role-based routing
├── contracts/                # shared API/type contracts — source of truth for cross-slice interfaces
└── docs/                     # SRS, ERD, working agreements, task breakdowns
```

The backend follows a layered architecture (handlers → services → repositories → middleware); state-transition rules are centralized in service files, not duplicated across handlers.

## API Overview

All endpoints are versioned under `/api/v1`, authenticated via `Authorization: Bearer <JWT>` except where marked public, and return a consistent envelope:

```json
{ "status": "success" | "fail", "data": "...", "message": "..." }
```

| Area | Example endpoints |
|---|---|
| Auth | `POST /auth/login`, `POST /auth/register`, `POST /auth/reset-password` |
| Attendance | `POST /attendance/clock-in`, `POST /attendance/sync`, `GET /attendance/me` |
| Tasks | `POST /tasks`, `PATCH /tasks/:id/status`, `POST /tasks/:id/timer/start` |
| Timesheets | `PUT /timesheets/:id/submit`, `PUT /timesheets/:id/approve` |
| Dashboard/Reports | `GET /dashboard/manager`, `GET /reports/attendance` |
| Admin | `POST /admin/users`, `POST /admin/teams` |
| Notifications | `POST /notifications/telegram/link`, `POST /telegram/webhook` |

Full request/response shapes, error codes, and the complete endpoint catalog live in `/contracts` and SRS Appendix B.

## Team & Ownership

Four developers, four vertical slices — each owns the full stack (database + Go API + React screens) for their area, not just one layer:

| Developer | Slice | Core tables |
|---|---|---|
| Dev 1 | Auth, Users, Organizations, Teams & Admin | `organizations`, `users`, `teams`, `team_members` |
| Dev 2 | Attendance & Offline Sync | `attendance_logs` |
| Dev 3 | Tasks & Work Timer | `tasks`, `task_assignments`, `task_time_logs` |
| Dev 4 | Timesheets, Dashboard, Reports & Notifications | `timesheets`, `notifications`, `telegram_subscribers` |

Dev 1's auth/RBAC middleware is a hard dependency for the other three slices and is built first. Dev 4's timesheet aggregation depends on stable data contracts from Dev 2 and Dev 3, published to `/contracts` before that work begins. Full ownership detail, task breakdowns, and working rules live in `/docs`.

## Development Workflow

- `main` is always protected and deployable — no direct pushes.
- Each developer works on their own feature branch (`feature/dev1-auth`, `feature/dev2-attendance`, etc.).
- All changes go through Pull Requests; at least one other developer reviews before merge.
- Schema changes are made exclusively via GORM migrations, named with a developer prefix or sequence number to avoid collisions. Never edit another developer's merged migration — write a new one.
- Shared interfaces (API shapes, types, table contracts) live in `/contracts` and are agreed on *before* implementation starts, not discovered mid-build.
- No unilateral changes to shared code — raise it in the group first.

**Definition of Done** for any task:
1. Code is written and works
2. Basic error handling is present
3. Matches the corresponding FR/NFR in the SRS
4. PR is reviewed and merged
5. Tracked task is marked complete

## Testing

Each developer owns the test suite for their slice's highest-risk path:

| Owner | Focus |
|---|---|
| Dev 1 | RBAC boundary + cross-organization isolation |
| Dev 2 | Full offline-to-online sync state machine, including tamper rejection |
| Dev 3 | Task/timer state machine — valid and invalid transitions |
| Dev 4 | Timesheet aggregation accuracy against known attendance + task-timer fixtures |

CI runs lint and the test suite on every pull request via GitHub Actions.

## Offline-First Design Notes

A few implementation details worth knowing before touching the attendance or task-timer code:

- Offline records are queued in Dexie.js (IndexedDB), encrypted at rest with AES-256, keys derived per-session and never persisted in plaintext.
- Every queued record carries a client-computed device hash; the server independently re-verifies it on sync — the client's word alone is never trusted.
- Sync is idempotent via a client-generated `record_uuid`, so a retried batch never double-counts.
- The app requests persistent storage (`navigator.storage.persist()`) to reduce the risk of browser eviction under storage pressure, particularly on iOS Safari, and offers a manual export of queued records as a fallback.
- A record that fails hash/timestamp verification is marked `REJECTED_TAMPERED`, excluded from attendance totals, and surfaced to the employee's Manager for review — it is never silently dropped.

## Roadmap

**Release 1.1 (post-MVP validation, pre-commercial launch):**
- Per-seat ETB subscription billing (Basic/Pro/Enterprise)
- Chapa and Telebirr payment integration
- Self-service plan management and invoicing

**Release 2.0+:**
- Automated payroll disbursement
- Client portal for international outsourcing clients
- Git/IDE proof-of-work integration
- Native mobile apps
- AI-assisted productivity analytics
- Hardened multi-org self-serve growth

## Documentation

- `SRS.pdf` — full Software Requirements Specification (IEEE 830-1998), functional/non-functional requirements, database schema, state machines, API surface, and traceability matrix
- `/contracts` — live source of truth for cross-slice API and data contracts

## License

INSA CYBER TALENT, Group 5. License terms to be finalized before any public or commercial release.

## Status
🚧 In active development.

| Name             | ID           | Class Room |
|------------------|--------------|------------|
| Amanuel Musa     | CTC-1441-26  | ROOM 1     |
| Amanuel Habtamu  | CTC-3561-26  | ROOM 1     |
| Abdurezak Anwar  | CTC-2844-26  | ROOM 1     |
| Aymen Muhammed   | CTC-1544-26  | ROOM 1     |
