package common

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ID struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
}

type Timestamps struct {
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type TenantScoped struct {
	OrgID uuid.UUID `gorm:"type:uuid;not null;index" json:"org_id"`
}

type SoftDeletable struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BaseModel combines ID, TenantScoped, and Timestamps for standard entities
type BaseModel struct {
	ID
	TenantScoped
	Timestamps
}
