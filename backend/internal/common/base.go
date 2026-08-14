package common

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel is embedded by every org-scoped entity.
type BaseModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrgID     uuid.UUID  `gorm:"type:uuid;not null;index"                        json:"org_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index"                                           json:"deleted_at,omitempty"`
}
