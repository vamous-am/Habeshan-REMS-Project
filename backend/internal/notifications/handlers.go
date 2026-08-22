package notifications

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// LinkTelegram handles POST /api/v1/notifications/telegram/link
// Body: { "chat_id": "123456789" }
// FR-NOTIFY-01
func (h *Handler) LinkTelegram(c *fiber.Ctx) error {
	// Parse request body
	var body struct {
		ChatID string `json:"chat_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	if body.ChatID == "" {
		return common.Fail(c, fiber.StatusBadRequest, "chat_id is required")
	}

	// TODO: once Dev 1's JWT middleware is wired in, pull userID from JWT claims
	// For now we accept it as a query param so the endpoint is testable as a shell
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "valid user_id query param required")
	}

	if err := h.service.LinkTelegram(userID, body.ChatID); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{
		"message": "Telegram account linked successfully",
	})
}

// GetTelegramStatus handles GET /api/v1/notifications/telegram/status
func (h *Handler) GetTelegramStatus(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "valid user_id query param required")
	}

	sub, err := h.service.GetSubscriber(userID)
	if err != nil {
		return common.HandleError(c, err)
	}

	if sub == nil {
		return common.OK(c, fiber.Map{"linked": false})
	}

	return common.OK(c, fiber.Map{
		"linked":    true,
		"chat_id":   sub.ChatID,
		"is_active": sub.IsActive,
		"linked_at": sub.LinkedAt,
	})
}
