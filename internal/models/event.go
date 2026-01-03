package models

import "time"

type AnalyticsEvent struct {
	ID         uint           `gorm:"primaryKey"`
	AppName    string         `gorm:"size:64;not null;index"`
	UserID     *string        `gorm:"index"`
	EventName  string         `gorm:"size:64;not null;index"`
	Properties map[string]any `gorm:"type:jsonb"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;index"`
}
