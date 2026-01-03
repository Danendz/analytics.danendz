package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"analytics-svc/internal/bus"
	"analytics-svc/internal/db"
	"analytics-svc/internal/ingest"
	"analytics-svc/internal/routes"
	"analytics-svc/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using OS env")
	}

	database := db.Connect()

	writer := ingest.NewWriter(database)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go writer.Run(ctx)

	trackService := services.TrackService{Writer: writer}

	if os.Getenv("RABBITMQ_URL") != "" {
		consumer := bus.NewRabbitConsumer(trackService)
		go consumer.ConsumeForever(ctx)
	}

	app := fiber.New()
	routes.Register(app, trackService)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
		_ = app.Shutdown()
	}()

	if os.Getenv("APP_PORT") != "" {
		consumer := bus.NewRabbitConsumer(trackService)
		go consumer.ConsumeForever(ctx)
	}

	log.Fatal(app.Listen(fmt.Sprintf(":%s", os.Getenv("APP_PORT"))))
}
