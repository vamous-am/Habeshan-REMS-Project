package common

import "github.com/gofiber/fiber/v2"

// Envelope is the standard API response shape.
// { "status": "success"|"fail", "data": ..., "message": "..." }
type Envelope struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Status: "success", Data: data})
}

func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Status: "success", Data: data})
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Envelope{Status: "fail", Message: message})
}
