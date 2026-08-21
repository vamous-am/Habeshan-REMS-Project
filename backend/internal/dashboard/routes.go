package dashboard

import "github.com/gofiber/fiber/v2"

// RegisterRoutes wires this feature's endpoints onto the app.
func RegisterRoutes(app *fiber.App) {
	app.Get("/api/v1/dashboard/manager", GetManagerDashboard)
}
