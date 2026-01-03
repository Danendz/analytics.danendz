package db

import (
	"log"
	"os"

	"analytics-svc/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := database.AutoMigrate(&models.AnalyticsEvent{}); err != nil {
		log.Fatal(err)
	}

	return database
}
