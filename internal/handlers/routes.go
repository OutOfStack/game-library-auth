package handlers

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers routes for auth service
func RegisterRoutes(app *fiber.App) {
	app.Post("/signin", signInHandler)
}
