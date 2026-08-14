# Frozen DB Schema Contracts

Cross-slice table shapes — field lists, types, and foreign keys.
**Do not change without team agreement.**

## organizations
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| name | varchar | |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

## users
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id |
| email | varchar | UNIQUE (org_id, email) WHERE deleted_at IS NULL |
| password_hash | varchar | bcrypt cost ≥ 10 |
| role | enum | admin \| manager \| employee |
| created_at | timestamptz | |
| updated_at | timestamptz | |
| deleted_at | timestamptz | soft delete |

## teams
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id |
| name | varchar | |
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
| org_id | uuid | FK → organizations.id |
| user_id | uuid | FK → users.id |
| clocked_in_at | timestamptz | |
| clocked_out_at | timestamptz | nullable |
| sync_status | enum | offline_logged \| pending_sync \| synced_verified \| rejected_tampered |
| device_hash | varchar | tamper-evident |
| record_uuid | uuid | idempotency key |
| created_at | timestamptz | |
| updated_at | timestamptz | |

## tasks
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id |
| title | varchar | |
| description | text | nullable |
| status | enum | to_do \| in_progress \| paused \| blocked \| completed |
| created_by | uuid | FK → users.id |
| created_at | timestamptz | |
| updated_at | timestamptz | |

## task_assignments
| Column | Type | Notes |
|---|---|---|
| task_id | uuid | FK → tasks.id |
| user_id | uuid | FK → users.id |
| PRIMARY KEY | (task_id, user_id) | |

## task_time_logs
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id |
| task_id | uuid | FK → tasks.id |
| user_id | uuid | FK → users.id |
| started_at | timestamptz | |
| stopped_at | timestamptz | nullable |
| pause_reason | varchar | nullable |
| sync_status | enum | same as attendance_logs |
| device_hash | varchar | |
| record_uuid | uuid | idempotency key |
| created_at | timestamptz | |
| updated_at | timestamptz | |

## timesheets
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id |
| user_id | uuid | FK → users.id |
| period_start | date | |
| period_end | date | |
| status | enum | draft \| submitted \| approved \| rejected |
| reviewed_by | uuid | FK → users.id, nullable |
| created_at | timestamptz | |
| updated_at | timestamptz | |

## notifications
| Column | Type | Notes |
|---|---|---|
| id | uuid | PK |
| org_id | uuid | FK → organizations.id |
| user_id | uuid | FK → users.id |
| message | text | |
| read | boolean | default false |
| created_at | timestamptz | |

## telegram_subscribers
| Column | Type | Notes |
|---|---|---|
| user_id | uuid | PK, FK → users.id |
| chat_id | varchar | Telegram chat ID |
| linked_at | timestamptz | |
