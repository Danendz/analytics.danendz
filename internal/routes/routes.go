package routes

import (
	"analytics-svc/internal/handlers"
	"analytics-svc/internal/services"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App, trackService services.TrackService) {
	api := app.Group("/api/analytics")

	api.Get("/health", func(c *fiber.Ctx) error { return c.SendString("ok") })

	h := handlers.HttpTrackHandler{Service: trackService}
	api.Post("/track", h.Track)
}
