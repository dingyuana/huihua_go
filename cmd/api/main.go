package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
			log.Printf("[ERROR] %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal server error",
			})
		},
	})

	// Global middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3002,http://localhost:3003,http://localhost:5173",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	app.Use(logger.New())
	app.Use(recover.New())

	// Setup routes
	setupRoutes(app, db, rdb, cfg)

	log.Fatal(app.Listen(":8080"))
}

func setupRoutes(app *fiber.App, db *database.DB, rdb *database.RedisClient, cfg *config.Config) {
	// Health check (public)
	app.Get("/health", handler.HealthCheck)

	// Auth routes (public)
	userRepo := repository.NewUserRepository(db.GetPool())
	authSvc := service.NewAuthService(userRepo, cfg)
	authHandler := handler.NewAuthHandler(authSvc)
	app.Post("/api/v1/auth/login", authHandler.Login)

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

	exchangeRateRepo := repository.NewExchangeRateRepository(db.GetPool())
	exchangeRateSvc := service.NewExchangeRateService(exchangeRateRepo)
	exchangeRateHandler := handler.NewExchangeRateHandler(exchangeRateSvc)
	api.Get("/exchange-rates", exchangeRateHandler.List)
	api.Post("/exchange-rates", exchangeRateHandler.Create)
	api.Get("/exchange-rates/:id", exchangeRateHandler.GetByID)
	api.Delete("/exchange-rates/:id", exchangeRateHandler.Delete)
	api.Get("/exchange-rates/convert", exchangeRateHandler.Convert)

	// Bank account routes
	bankRepo := repository.NewBankRepository(db.GetPool())
	bankSvc := service.NewBankService(bankRepo, accountRepo)
	bankHandler := handler.NewBankHandler(bankSvc)
	api.Get("/bank-accounts", bankHandler.List)
	api.Post("/bank-accounts", bankHandler.Create)
	api.Put("/bank-accounts/:id", bankHandler.Update)
	api.Delete("/bank-accounts/:id", bankHandler.Delete)
	api.Post("/bank-accounts/:id/adjust-balance", bankHandler.AdjustBalance)
	api.Get("/bank-accounts/:id/balance-adjustments", bankHandler.ListBalanceAdjustments)

	// Party routes
	partyRepo := repository.NewPartyRepository(db.GetPool())
	partySvc := service.NewPartyService(partyRepo)
	partyHandler := handler.NewPartyHandler(partySvc)
	api.Get("/parties", partyHandler.List)
	api.Post("/parties/import", partyHandler.ImportExcel)
	api.Get("/parties/:id", partyHandler.GetByID)
	api.Post("/parties", partyHandler.Create)
	api.Put("/parties/:id", partyHandler.Update)
	api.Delete("/parties/:id", partyHandler.Delete)

	// Account setup routes
	companyRepo := repository.NewCompanyRepository(db.GetPool())
	periodRepo := repository.NewPeriodRepository(db.GetPool())
	setupSvc := service.NewSetupService(companyRepo, periodRepo, accountSvc, db.GetPool())
	setupHandler := handler.NewSetupHandler(setupSvc)
	api.Get("/account-setup/status", setupHandler.GetStatus)
	api.Post("/account-setup/wizard", setupHandler.CreateCompany)

	// Asset Depreciation routes
	depreciationRepo := repository.NewAssetDepreciationRepository(db.GetPool())
	journalRepo := repository.NewJournalRepository(db.GetPool())
	glEntryRepo := repository.NewGLEntryRepository(db.GetPool())
	depreciationSvc := service.NewAssetDepreciationService(depreciationRepo, journalRepo)
	depreciationHandler := handler.NewAssetDepreciationHandler(depreciationSvc)
	api.Post("/assets/:id/depreciation/schedule", depreciationHandler.CreateSchedule)
	api.Get("/assets/:id/depreciation/schedule", depreciationHandler.GetSchedule)
	api.Post("/depreciation/run", depreciationHandler.RunDepreciation)
	api.Get("/depreciation/run", depreciationHandler.ListDepreciationRuns)
	api.Put("/assets/:id/depreciation", depreciationHandler.CreateSchedule) // alias for schedule creation

	// Invoice routes
	invoiceRepo := repository.NewInvoiceRepository(db.GetPool())
	invoiceSvc := service.NewInvoiceService(invoiceRepo)
	invoiceHandler := handler.NewInvoiceHandler(invoiceSvc)
	api.Get("/invoices", invoiceHandler.List)
	api.Post("/invoices", invoiceHandler.Create)
	api.Post("/invoices/import", invoiceHandler.ImportFromExcel)
	api.Post("/invoices/import-excel", invoiceHandler.ImportExcelFile)
	api.Post("/invoices/parse", invoiceHandler.Parse)
	api.Get("/invoices/:id", invoiceHandler.GetByID)
	api.Put("/invoices/:id/status", invoiceHandler.UpdateStatus)

	// Classification rule routes
	classificationRuleRepo := repository.NewClassificationRuleRepository(db.GetPool())
	classificationRuleSvc := service.NewClassificationRuleService(classificationRuleRepo)
	classificationRuleHandler := handler.NewClassificationRuleHandler(classificationRuleSvc)
	api.Get("/classification-rules", classificationRuleHandler.List)
	api.Post("/classification-rules", classificationRuleHandler.Create)
	api.Put("/classification-rules/:id", classificationRuleHandler.Update)
	api.Delete("/classification-rules/:id", classificationRuleHandler.Delete)
	api.Post("/classification-rules/reorder", classificationRuleHandler.Reorder)
	api.Post("/classification-rules/match", classificationRuleHandler.Match)
	api.Post("/classification-rules/seed", classificationRuleHandler.Seed)

	// Bank transaction routes
	bankTransactionRepo := repository.NewBankTransactionRepository(db.GetPool())
	bankTxnSvc := service.NewBankTransactionService(bankTransactionRepo, classificationRuleSvc, bankRepo)
	bankTxnHandler := handler.NewBankTransactionHandler(bankTxnSvc)
	api.Get("/bank-transactions", bankTxnHandler.List)
	api.Post("/bank-transactions/preview", bankTxnHandler.PreviewExcel)
	api.Post("/bank-transactions/import", bankTxnHandler.Import)
	api.Post("/bank-transactions/classify-all", bankTxnHandler.ClassifyAll)
	api.Post("/bank-transactions/:id/classify", bankTxnHandler.Classify)
	api.Post("/bank-transactions/:id/mark-matched", bankTxnHandler.MarkMatched)
	api.Get("/bank-transactions/unmatched", bankTxnHandler.GetUnmatched)
	api.Get("/bank-transactions/:id", bankTxnHandler.GetByID)
	api.Delete("/bank-transactions/:id", bankTxnHandler.Delete)

	// Voucher template routes
	voucherTemplateRepo := repository.NewVoucherTemplateRepository(db.GetPool())
	voucherTemplateSvc := service.NewVoucherTemplateService(voucherTemplateRepo, accountRepo)
	voucherTemplateHandler := handler.NewVoucherTemplateHandler(voucherTemplateSvc)
	api.Get("/voucher-templates", voucherTemplateHandler.List)
	api.Post("/voucher-templates", voucherTemplateHandler.Create)
	api.Get("/voucher-templates/:id", voucherTemplateHandler.GetByID)
	api.Put("/voucher-templates/:id", voucherTemplateHandler.Update)
	api.Delete("/voucher-templates/:id", voucherTemplateHandler.Delete)
	api.Get("/voucher-templates/numbering-rule", voucherTemplateHandler.GetNumberingRule)
	api.Post("/voucher-templates/numbering-rule", voucherTemplateHandler.UpdateNumberingRule)
	api.Post("/voucher-templates/numbering-rule/next", voucherTemplateHandler.GenerateNextNumber)

	// Voucher state machine routes
	auditRepo = repository.NewAuditRepository(db.GetPool())
	stateMachine := service.NewVoucherStateMachine(journalRepo, auditRepo, glEntryRepo)
	voucherSvc := service.NewVoucherService(journalRepo, voucherTemplateSvc)
	voucherHandler := handler.NewVoucherHandler(stateMachine, journalRepo, voucherSvc)
	api.Get("/vouchers", voucherHandler.List)
	api.Post("/vouchers", voucherHandler.Create)
	api.Get("/vouchers/:id", voucherHandler.GetByID)
	api.Put("/vouchers/:id", voucherHandler.Update)
	api.Delete("/vouchers/:id", voucherHandler.Delete)
	api.Post("/vouchers/:id/submit", voucherHandler.Submit)
	api.Post("/vouchers/:id/approve", voucherHandler.Approve)
	api.Post("/vouchers/:id/reject", voucherHandler.Reject)
	api.Post("/vouchers/:id/cancel", voucherHandler.Cancel)
	api.Post("/vouchers/:id/reverse", voucherHandler.Reverse)
	api.Get("/vouchers/:id/status", voucherHandler.GetStatus)
	api.Get("/vouchers/:id/transitions", voucherHandler.GetTransitions)

	// Opening balance routes
	obRepo := repository.NewOpeningBalanceRepository(db.GetPool())
	obSvc := service.NewOpeningBalanceService(obRepo, accountRepo)
	obHandler := handler.NewOpeningBalanceHandler(obSvc)
	api.Post("/opening-balances/import", obHandler.Import)
	api.Get("/opening-balances", obHandler.List)
	api.Get("/opening-balances/trial-balance", obHandler.GetTrialBalance)
	api.Post("/opening-balances/validate", obHandler.Validate)
	api.Get("/opening-balances/:account_id", obHandler.GetByAccount)

	// Accounting period routes
	periodRepo = repository.NewPeriodRepository(db.GetPool())
	periodSvc := service.NewPeriodService(periodRepo, journalRepo, glEntryRepo, accountRepo, depreciationRepo)
	periodHandler := handler.NewPeriodHandler(periodSvc)
	api.Get("/periods", periodHandler.List)
	api.Get("/periods/current", periodHandler.GetCurrent)
	api.Get("/periods/voucher-gaps", periodHandler.VoucherGaps)
	api.Get("/periods/pre-close-check", periodHandler.PreCloseCheck)
	api.Post("/periods/:period_no/close", periodHandler.Close)
	api.Post("/periods/:period_no/unclose", periodHandler.Unclose)

	// Reconciliation (核销) routes
	reconRepo := repository.NewReconciliationRepository(db.GetPool())
	reconciliationSvc := service.NewReconciliationService(db.GetPool(), bankTransactionRepo, invoiceRepo, reconRepo, journalRepo)
	reconciliationHandler := handler.NewReconciliationHandler(reconciliationSvc)
	api.Post("/reconciliation/run", reconciliationHandler.Run)
	api.Get("/reconciliation/pairs", reconciliationHandler.ListPairs)
	api.Post("/reconciliation/pairs/:id/confirm", reconciliationHandler.ConfirmPair)
	api.Post("/reconciliation/pairs/:id/unconfirm", reconciliationHandler.UnconfirmPair)
	api.Get("/reconciliation/unmatched", reconciliationHandler.GetUnmatched)
	api.Post("/reconciliation/manual", reconciliationHandler.ManualMatch)

	// Bank reconciliation (银企对账) routes
	bankReconciliationSvc := service.NewBankReconciliationService(bankTransactionRepo, journalRepo, bankRepo, glEntryRepo)
	reconHandler := handler.NewBankReconciliationHandler(bankReconciliationSvc)

	// Bank reconciliation routes
	api.Post("/bank-reconciliation/reconcile", reconHandler.Reconcile)
	api.Get("/bank-reconciliation/report", reconHandler.GetReport)
	api.Get("/bank-reconciliation/balance-check", reconHandler.BalanceCheck)
	api.Get("/bank-reconciliation/diff-report", reconHandler.GetDiffReport)
	api.Post("/bank-reconciliation/mark-done", reconHandler.MarkDone)
	api.Get("/bank-reconciliation/status", reconHandler.GetStatus)

	// Financial report routes
	reportSvc := service.NewReportService(glEntryRepo, obRepo, accountRepo, periodRepo)
	reportHandler := handler.NewReportHandler(reportSvc)
	api.Get("/reports/trial-balance", reportHandler.GetTrialBalance)
	api.Get("/reports/income-statement", reportHandler.GetIncomeStatement)
	api.Get("/reports/balance-sheet", reportHandler.GetBalanceSheet)
	api.Get("/reports/cash-flow", reportHandler.GetCashFlowStatement)

	// Approval workflow routes
	approvalRepo := repository.NewApprovalRepository(db.GetPool())
	approvalSvc := service.NewApprovalService(approvalRepo, journalRepo)
	approvalHandler := handler.NewApprovalHandler(approvalSvc)
	api.Post("/approvals/submit", approvalHandler.SubmitForApproval)
	api.Post("/approvals/:id/approve", approvalHandler.Approve)
	api.Post("/approvals/:id/reject", approvalHandler.Reject)
	api.Get("/approvals/pending", approvalHandler.GetPendingTasks)
	api.Get("/approvals/history", approvalHandler.GetApprovalHistory)
	api.Get("/approvals/voucher/:id/status", approvalHandler.GetVoucherApprovalStatus)
	api.Post("/approval-flows", approvalHandler.CreateApprovalFlow)
	api.Get("/approval-flows", approvalHandler.ListApprovalFlows)
	api.Put("/approval-flows/:id", approvalHandler.UpdateApprovalFlow)
	api.Delete("/approval-flows/:id", approvalHandler.DeleteApprovalFlow)

	// Voucher auto-generate routes (from bank transactions)
	// Placed after approvalSvc is initialized so it can be injected
	autoGenSvc := service.NewVoucherAutoGenerateService(
		journalRepo, glEntryRepo, bankTransactionRepo,
		invoiceRepo, accountRepo, classificationRuleSvc, voucherTemplateSvc, approvalSvc,
	)
	// Wire auto-gen service into bank transaction handler for post-import voucher auto-generation
	bankTxnHandler.InjectAutoGenSvc(autoGenSvc)
	autoGenHandler := handler.NewVoucherAutoGenerateHandler(autoGenSvc)
	api.Post("/bank-transactions/:id/generate-voucher", autoGenHandler.GenerateFromBankTxn)
	api.Post("/bank-transactions/batch-generate", autoGenHandler.BatchGenerate)
	api.Post("/invoices/:id/generate-voucher", autoGenHandler.GenerateFromInvoice)

	// Payment entry routes (收款/付款单)
	paymentRepo := repository.NewPaymentEntryRepository(db.GetPool())
	paymentSvc := service.NewPaymentEntryService(paymentRepo, partyRepo, bankRepo, accountRepo)
	paymentHandler := handler.NewPaymentEntryHandler(paymentSvc, bankTransactionRepo)
	api.Get("/payment-entries", paymentHandler.List)
	api.Post("/payment-entries", paymentHandler.CreateFromBankTransaction)
	api.Get("/payment-entries/:id", paymentHandler.GetByID)
	api.Put("/payment-entries/:id", paymentHandler.Update)
	api.Delete("/payment-entries/:id", paymentHandler.Delete)
}
