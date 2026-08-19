package notifications

import (
	"time"

	"github.com/google/uuid"
)

// Notification — NOT embedding common.BaseModel (no updated_at/deleted_at
// in the frozen schema).
type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrgID     uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Type      string    `gorm:"type:varchar(50);not null"` // e.g. timesheet_approved, timesheet_rejected, clock_in_reminder
	Message   string    `gorm:"type:text;not null"`
	Read      bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (Notification) TableName() string {
	return "notifications"
}

// TelegramSubscriber — also NOT embedding BaseModel. user_id is the PK,
// no id/org_id/deleted_at at all.
type TelegramSubscriber struct {
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	ChatID   string    `gorm:"type:varchar(50);not null"`
	IsActive bool      `gorm:"not null;default:true"`
	LinkedAt time.Time `gorm:"not null;default:now()"`
}

func (TelegramSubscriber) TableName() string {
	return "telegram_subscribers"
}
