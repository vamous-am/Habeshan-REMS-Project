package timesheets

import (
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
)

type Timesheet struct {
	common.ID
	common.TenantScoped
	common.Timestamps

	UserID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	PeriodStart     time.Time  `gorm:"type:date;not null" json:"period_start"`
	PeriodEnd       time.Time  `gorm:"type:date;not null" json:"period_end"`
	TotalHours      float64    `gorm:"type:numeric(6,2);not null" json:"total_hours"`
	Status          string     `gorm:"type:varchar(20);not null;default:'draft';check:status IN ('draft','submitted','approved','rejected')" json:"status"`
	ReviewedBy      *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	RejectionReason *string    `gorm:"type:text" json:"rejection_reason,omitempty"`

	Organization auth.Organization `gorm:"foreignKey:OrgID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	User         auth.User         `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Reviewer     *auth.User        `gorm:"foreignKey:ReviewedBy;references:ID;constraint:OnDelete:SET NULL" json:"-"`
}

func (Timesheet) TableName() string {
	return "timesheets"
}
