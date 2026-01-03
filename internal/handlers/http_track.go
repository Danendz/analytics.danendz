package handlers

import (
	"analytics-svc/internal/services"
	"errors"

	"github.com/gofiber/fiber/v2"
)

type TrackRequest struct {
	EventId    string         `json:"event_id"`
	AppName    string         `json:"app_name"`
	UserID     *string        `json:"user_id"`
	EventName  string         `json:"event_name"`
	Properties map[string]any `json:"properties"`
}

type HttpTrackHandler struct {
	Service services.TrackService
}

func (h HttpTrackHandler) Track(c *fiber.Ctx) error {
	var req TrackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_json"})
	}

	_, err := h.Service.Track(services.TrackInput{
		EventId:    req.EventId,
		AppName:    req.AppName,
		UserID:     req.UserID,
		EventName:  req.EventName,
		Properties: req.Properties,
	})

	if errors.Is(err, services.ErrInvalid) {
		return c.Status(422).JSON(fiber.Map{"error": "app_name and event_name required"})
	}

	if errors.Is(err, services.ErrQueueFull) {
		return c.Status(429).JSON(fiber.Map{"error": "queue_full"})
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal"})
	}

	return c.Status(202).JSON(fiber.Map{"status": "queued"})
}
