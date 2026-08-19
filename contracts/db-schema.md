
# Frozen DB Schema Contracts

Cross-slice table shapes — field lists, types, and foreign keys.
**Do not change without team agreement.**

This intentionally deviates from the SRS's literal `INT AUTO_INCREMENT` / `DATETIME` schema — see "Deviation from SRS" at the bottom of this file.

Column notes ending in an FR/BR/C-number cite the exact SRS requirement that field satisfies — added during the Sprint 0 contract-vs-SRS reconciliation pass, so nobody has to re-derive why a field exists.

**Soft delete is NOT universal.** Only `organizations` and `users` have a `deleted_at` column — matching FR-AUTH-09, which only requires soft-deleting user accounts. Every other table uses hard deletes (or, where relevant, a status field like `sync_status`) instead. Check each table's own column list below — don't assume `deleted_at` exists unless it's listed.

**`org_id` is present on tables that need direct org-scoped queries, and absent on the two pure junction tables.** `attendance_logs`, `tasks`, `task_time_logs`, `timesheets`, and `notifications` all carry `org_id` directly — this is a deliberate denormalization beyond what the SRS specifies (the SRS scopes these transitively through `user_id`/`task_id` instead). The reason: every one of these tables is queried directly and filtered by organization constantly (attendance reports, task boards, timesheet approval queues, notification feeds), so a direct `org_id` column avoids a join on every single request and makes the org-scoping required by NFR-SEC-03 impossible to accidentally skip in a query. `team_members` and `task_assignments` are pure junction tables — every query against them already goes through `teams`/`tasks` or `users`, which are themselves org-scoped, so a redundant `org_id` on the junction row adds storage and write overhead with no query benefit. That's the rule: **direct `org_id` on tables queried independently; omitted on junction tables whose parent is already org-scoped.**

## organizations
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| name | varchar(150) | NOT NULL |
| currency | varchar(3) | default `ETB` — FR-ADMIN-04 |
| timezone | varchar(50) | default `Africa/Addis_Ababa` — FR-ADMIN-04 |
| seat_count | int | default 0 — required by constraint C8 (schema must be complete now even though billing is Release 1.1) |
| subscription_status | enum | `trial \| active \| suspended`, default `trial` — C8 |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

## users
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id |
| email | varchar(255) | UNIQUE (org_id, email) WHERE deleted_at IS NULL |
| password_hash | varchar | bcrypt cost ≥ 10 — FR-AUTH-02 |
| full_name | varchar(100) | NOT NULL — registration API, Appendix B.2 |
| phone | varchar(20) | nullable — registration API |
| role | enum | admin \| manager \| employee |
| status | enum | `active \| inactive`, default `active` — FR-ADMIN-01 (deactivate is reversible, distinct from soft delete) |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | soft delete — FR-AUTH-09 |

## teams
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id |
| name | varchar(100) | |
| manager_id | uuid | FK → users.id |
| created_at | timestamptz | |
| updated_at | timestamptz | |

## team_members
| Column | Type | Notes |
|---|---|---|
| team_id | uuid | FK → teams.id |
| user_id | uuid | FK → users.id |
| PRIMARY KEY | (team_id, user_id) | |

## attendance_logs
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id — denormalized from SRS (SRS scopes transitively via user_id only); kept deliberately so every query is trivially org-scoped, directly serving NFR-SEC-03 |
| user_id | uuid | FK → users.id |
| clocked_in_at | timestamptz | NOT NULL |
| clocked_out_at | timestamptz | nullable |
| total_hours | numeric(6,2) | nullable, computed on clock-out — needed for FR-DASH-02 reporting |
| sync_status | enum | offline_logged \| pending_sync \| synced_verified \| rejected_tampered |
| device_hash | varchar(128) | tamper-evident — FR-ATT-04 |
| record_uuid | uuid | UNIQUE, idempotency key — FR-ATT-08 |
| created_at | timestamptz | |
| updated_at | timestamptz | |

## tasks
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id |
| title | varchar(150) | NOT NULL |
| description | text | nullable |
| priority | enum | `low \| medium \| high` — default `medium`. Three values only, not four — FR-TASK-01 |
| due_date | date | nullable — FR-TASK-01. Plain `date`, not timestamptz (matches SRS; no time-of-day needed) |
| status | enum | to_do \| in_progress \| paused \| blocked \| completed |
| created_by | uuid | FK → users.id |
| created_at | timestamptz | |
| updated_at | timestamptz | |

## task_assignments
| Column | Type | Notes |
|---|---|---|
| task_id | uuid | FK → tasks.id |
| user_id | uuid | FK → users.id |
| assigned_at | timestamptz | default now — SRS Table 7 |
| PRIMARY KEY | (task_id, user_id) | |

## task_time_logs
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id — denormalized, same rationale as attendance_logs |
| task_id | uuid | FK → tasks.id |
| user_id | uuid | FK → users.id |
| started_at | timestamptz | NOT NULL |
| stopped_at | timestamptz | nullable |
| duration_minutes | int | nullable, computed on stop — SRS Table 8 |
| pause_reason | varchar(50) | nullable |
| sync_status | enum | same as attendance_logs |
| device_hash | varchar(128) | |
| record_uuid | uuid | idempotency key — not listed in SRS Table 8, but required by FR-TASK-10 ("same tamper-evident hashing and synchronization state machine as attendance records," which includes idempotent sync per FR-ATT-08). Treated as an SRS omission, kept in contract. |
| created_at | timestamptz | |
| updated_at | timestamptz | |

## timesheets
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id — denormalized, same rationale as attendance_logs |
| user_id | uuid | FK → users.id |
| period_start | date | |
| period_end | date | |
| total_hours | numeric(6,2) | NOT NULL, auto-computed — FR-TS-01 |
| status | enum | draft \| submitted \| approved \| rejected |
| reviewed_by | uuid | FK → users.id, nullable — not in SRS table explicitly, kept as a reasonable addition (records which Manager acted, supports BR-07's audit intent) |
| rejection_reason | text | nullable, required when status = rejected — FR-TS-04 |
| created_at | timestamptz | |
| updated_at | timestamptz | |

## notifications
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id — denormalized, same rationale as attendance_logs |
| user_id | uuid | FK → users.id |
| type | varchar(50) | e.g. `timesheet_approved`, `timesheet_rejected`, `clock_in_reminder` — needed to distinguish notification kinds for FR-NOTIFY-02/03/04 |
| message | text | |
| read | boolean | default false |
| created_at | timestamptz | |

## telegram_subscribers
| Column | Type | Notes |
|---|---|---|
| user_id | uuid | PK, FK → users.id |
| chat_id | varchar(50) | Telegram chat ID |
| is_active | boolean | default true — SRS Table 11, lets a user disable notifications without unlinking |
| linked_at | timestamptz | not in SRS, kept as a harmless addition |

---

## Deviation from SRS

The SRS's schema appendix (Section 3.4) specifies auto-incrementing integer PKs and `DATETIME` columns, and constraint C8 says the schema must be implemented "exactly as specified." This project instead uses UUID primary keys, `timestamptz`, and mandatory `org_id` scoping on every table, per the shared `BaseModel` struct established in Sprint 0 (`internal/common/base.go`).

This is a structural/implementation choice, not a change to any functional requirement — every FR is still fully satisfied by this schema, and the response envelope, error codes, and API surface all match the SRS as written. Rationale: UUIDs avoid PK collisions across a multi-tenant system without a central sequence, and `org_id` at the struct level makes tenant isolation (NFR-SEC-03) harder to accidentally miss than relying on transitive joins.
