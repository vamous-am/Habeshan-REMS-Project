package notifications

import (
	"time"

	"github.com/google/uuid"

	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
)

type Notification struct {
	common.ID
	common.TenantScoped

	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Read      bool      `gorm:"not null;default:false" json:"read"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`

	Organization auth.Organization `gorm:"foreignKey:OrgID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	User         auth.User         `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}

type TelegramSubscriber struct {
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	ChatID   string    `gorm:"type:varchar(50);not null" json:"chat_id"`
	IsActive bool      `gorm:"not null;default:true" json:"is_active"`
	LinkedAt time.Time `gorm:"not null;default:now()" json:"linked_at"`

	User auth.User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (TelegramSubscriber) TableName() string {
	return "telegram_subscribers"
}
