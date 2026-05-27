package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"huihua/finance/internal/config"
	"huihua/finance/internal/handler"
	"huihua/finance/internal/middleware"
	"huihua/finance/pkg/database"
)

func main() {
	// Load config
	cfg := config.Load()

	// Init DB
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	// Init Redis
	rdb := database.NewRedis(cfg)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Global middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Setup routes
	setupRoutes(app, db, rdb, cfg)

	log.Fatal(app.Listen(":8080"))
}

func setupRoutes(app *fiber.App, db *database.DB, rdb *database.RedisClient, cfg *config.Config) {
	// Health check (public)
	app.Get("/health", handler.HealthCheck)

	// Protected routes
	api := app.Group("/api/v1", middleware.Auth(cfg))
	api.Use(middleware.Tenant(db))

	// TODO: Add business handlers here
}