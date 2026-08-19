package timesheets

import (
	"time"

	"github.com/google/uuid"
)

// Timesheet — explicit fields, NOT embedding common.BaseModel, because the
// frozen schema has no deleted_at column for this table.
type Timesheet struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrgID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index"`
	PeriodStart     time.Time  `gorm:"type:date;not null"`
	PeriodEnd       time.Time  `gorm:"type:date;not null"`
	TotalHours      float64    `gorm:"type:numeric(6,2);not null"`
	Status          string     `gorm:"type:varchar(20);not null;default:'draft'"` // draft | submitted | approved | rejected
	ReviewedBy      *uuid.UUID `gorm:"type:uuid"`
	RejectionReason *string    `gorm:"type:text"`
	CreatedAt       time.Time  `gorm:"not null;default:now()"`
	UpdatedAt       time.Time  `gorm:"not null;default:now()"`
}

func (Timesheet) TableName() string {
	return "timesheets"
}
