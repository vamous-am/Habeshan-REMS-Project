package migrations

import (
	"gorm.io/gorm"

	"github.com/habeshan-rems/backend/internal/notifications"
)

func Migrate0010TelegramSubscribers(db *gorm.DB) error {
	return db.AutoMigrate(&notifications.TelegramSubscriber{})
}
