package common

import (
	"errors"
	"github.com/gofiber/fiber/v2"
)

// Envelope is the standard API response shape.
// { "status": "success"|"fail", "data": ..., "message": "..." }
type Envelope struct {
	Status  string `json:"status"` // "success" | "fail"
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{
		Status: "success",
		Data:   data,
	})
}

func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{
		Status: "success",
		Data:   data,
	})
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Envelope{
		Status:  "fail",
		Message: message,
	})
}

// ─────────────────────────────────────────────────────────────
// Sentinel errors — return these from services and repositories.
// ──────────────────────────────────────
var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource already exists")
	ErrInternal     = errors.New("internal server error")
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

// errStatus is the single place that maps sentinel errors → HTTP status + message.
var errStatus = map[error]struct {
	Status  int
	Message string
}{
	ErrNotFound:     {fiber.StatusNotFound, "resource not found"},
	ErrUnauthorized: {fiber.StatusUnauthorized, "unauthorized"},
	ErrForbidden:    {fiber.StatusForbidden, "forbidden"},
	ErrConflict:     {fiber.StatusConflict, "resource already exists"},
	ErrBadRequest:   {fiber.StatusBadRequest, "bad request"},
	ErrInternal:     {fiber.StatusInternalServerError, "internal server error"},
}

// HandleError converts a sentinel error into the correct HTTP response.
// Unrecognized errors fall back to 500 and never leak internal details.
func HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	for sentinel, mapped := range errStatus {
		if errors.Is(err, sentinel) {
			return Fail(c, mapped.Status, mapped.Message)
		}
	}

	// TODO: log the real error here once a logger is added
	return Fail(c, fiber.StatusInternalServerError, "internal server error")
}
