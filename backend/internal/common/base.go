package common

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ID is embedded by every table that has its own surrogate UUID primary key.
// Tables whose PK is something else (composite junction keys, or a FK
// promoted to PK like telegram_subscribers.user_id) do NOT embed this.
type ID struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
}

// Timestamps is embedded by every table that tracks both creation and
// modification time. Write-once tables (e.g. notifications) embed only
// CreatedAt directly instead — see that model for the pattern.
type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantScoped is embedded ONLY by tables queried directly and filtered by
// organization: users, teams, attendance_logs, tasks, task_time_logs,
// timesheets, notifications. it is deliberately omitted
// from team_members, task_assignments, and telegram_subscribers — those
// inherit tenant scope transitively through their parent/owning table.
type TenantScoped struct {
	OrgID uuid.UUID `gorm:"type:uuid;not null;index" json:"org_id"`
}

// SoftDeletable is embedded ONLY by organizations and users.
// §3.4, every other table is immutable/write-once history and must NOT
// have this — do not add it to a model without confirming against
// contracts/db-schema.md first.
type SoftDeletable struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
