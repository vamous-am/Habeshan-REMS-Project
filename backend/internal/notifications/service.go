package notifications

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// LinkTelegram links a Telegram chat ID to a user account.
// If the user already has a subscriber row, it updates it instead of inserting a new one.
func (s *Service) LinkTelegram(userID uuid.UUID, chatID string) error {
	sub := TelegramSubscriber{
		UserID:   userID,
		ChatID:   chatID,
		IsActive: true,
	}

	// FirstOrCreate: if a row with this UserID already exists, just fetch it.
	// If not, insert a new one. Then we Save to update ChatID + IsActive either way.
	result := s.db.Where(TelegramSubscriber{UserID: userID}).FirstOrCreate(&sub)
	if result.Error != nil {
		return result.Error
	}

	// Update in case they're re-linking with a new chat ID
	sub.ChatID = chatID
	sub.IsActive = true
	return s.db.Save(&sub).Error
}

// GetSubscriber fetches a user's Telegram subscriber row.
// Returns nil, nil if the user has never linked Telegram.
func (s *Service) GetSubscriber(userID uuid.UUID) (*TelegramSubscriber, error) {
	var sub TelegramSubscriber
	result := s.db.Where("user_id = ?", userID).First(&sub)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &sub, result.Error
}
