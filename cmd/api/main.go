package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"huihua/finance/internal/config"
	"huihua/finance/internal/handler"
	"huihua/finance/internal/middleware"
	"huihua/finance/internal/repository"
	"huihua/finance/internal/service"
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

	// Audit log routes
	auditRepo := repository.NewAuditRepository(db.GetPool())
	auditSvc := service.NewAuditService(auditRepo)
	auditHandler := handler.NewAuditHandler(auditSvc)

	api.Get("/audit-logs", auditHandler.ListAuditLogs)
	api.Get("/audit-logs/:object_type/:object_id", auditHandler.GetAuditLogsByObject)

	// Account routes
	accountRepo := repository.NewAccountRepository(db.GetPool())
	accountSvc := service.NewAccountService(accountRepo, db.GetPool())
	accountHandler := handler.NewAccountHandler(accountSvc)
	api.Get("/accounts/tree", accountHandler.GetTree)
	api.Post("/accounts/init-seed", accountHandler.InitFromSeed)

	// Bank account routes
	bankRepo := repository.NewBankRepository(db.GetPool())
	bankSvc := service.NewBankService(bankRepo, accountRepo)
	bankHandler := handler.NewBankHandler(bankSvc)
	api.Get("/bank-accounts", bankHandler.List)
	api.Post("/bank-accounts", bankHandler.Create)

	// Party routes
	partyRepo := repository.NewPartyRepository(db.GetPool())
	partySvc := service.NewPartyService(partyRepo)
	partyHandler := handler.NewPartyHandler(partySvc)
	api.Get("/parties", partyHandler.List)
	api.Post("/parties/import", partyHandler.ImportExcel)

	// Account setup routes
	companyRepo := repository.NewCompanyRepository(db.GetPool())
	periodRepo := repository.NewPeriodRepository(db.GetPool())
	setupSvc := service.NewSetupService(companyRepo, periodRepo, accountSvc)
	setupHandler := handler.NewSetupHandler(setupSvc)
	api.Get("/account-setup/status", setupHandler.GetStatus)
	api.Post("/account-setup/wizard", setupHandler.CreateCompany)
}