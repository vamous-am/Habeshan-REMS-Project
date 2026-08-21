package notifications

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	tg := app.Group("/api/v1/notifications/telegram")
	tg.Post("/link", handler.LinkTelegram)
	tg.Get("/status", handler.GetTelegramStatus)
}
