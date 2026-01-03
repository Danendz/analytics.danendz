package services

import (
	"errors"
	"time"

	"analytics-svc/internal/ingest"
	"analytics-svc/internal/models"
)

var ErrInvalid = errors.New("invalid event")
var ErrQueueFull = errors.New("queue full")

type TrackInput struct {
	AppName    string
	UserID     *string
	EventName  string
	Properties map[string]any
	At         time.Time
}

type TrackService struct {
	Writer *ingest.Writer
}

func (s TrackService) Track(in TrackInput) (*models.AnalyticsEvent, error) {
	if in.AppName == "" || in.EventName == "" {
		return nil, ErrInvalid
	}

	if in.At.IsZero() {
		in.At = time.Now().UTC()
	}

	ev := models.AnalyticsEvent{
		AppName:    in.AppName,
		UserID:     in.UserID,
		EventName:  in.EventName,
		Properties: in.Properties,
		CreatedAt:  in.At,
	}

	if ok := s.Writer.Enqueue(ev); !ok {
		return nil, ErrQueueFull
	}

	return &ev, nil
}
