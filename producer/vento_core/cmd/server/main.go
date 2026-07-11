package main

import (
	"context"
	"os"

	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/postgres"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/auth"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/config"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/workers"
)

func main() {
	// ── Load Config ──────────────────────────────────────────────
	cfg := config.Load()
	logger.Init(cfg.Env)

	// ── Infrastructure: Database ─────────────────────────────────
	db, err := postgres.NewConnection(cfg.DSN())
	if err != nil {
		logger.L().Error("❌ Database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// ── Infrastructure: JWT ──────────────────────────────────────
	jwtService := auth.NewJWTService(cfg.JWTSecret)

	// ── Infrastructure: Repositories ─────────────────────────────
	userRepo := postgres.NewPostgresUserRepository(db)
	productRepo := postgres.NewPostgresProductRepository(db)
	insumoRepo := postgres.NewPostgresInsumoRepository(db)
	paymentRepo := postgres.NewPostgresPaymentRepository(db)
	orderRepo := postgres.NewPostgresOrderRepository(db)
	metaRepo := postgres.NewMetaRepo(db)
	businessRepo := postgres.NewBusinessRepo(db)
	syncJobRepo := postgres.NewPostgresSyncJobRepository(db)
	tagRepo := postgres.NewPostgresTagRepository(db)
	serviceRepo := postgres.NewPostgresServiceRepository(db)
	clientRepo := postgres.NewPostgresClientRepository(db)
	expenseRepo := postgres.NewPostgresExpenseRepository(db)
	locationRepo := postgres.NewPostgresLocationRepository(db)

	// ── Application: Use Cases ───────────────────────────────────
	registerUC := usecase.NewRegisterUser(userRepo, jwtService, locationRepo)
	loginUC := usecase.NewLoginUser(userRepo, jwtService)
	resetPasswordUC := usecase.NewResetPasswordUsecases(userRepo)
	catalogUC := usecase.NewCatalogUsecases(productRepo, tagRepo, locationRepo)
	insumoUC := usecase.NewInsumoUsecases(insumoRepo, tagRepo, locationRepo)
	serviceUC := usecase.NewServiceUsecases(serviceRepo, tagRepo, insumoRepo)
	tagUC := usecase.NewTagUsecases(tagRepo)
	searchProductsUC := usecase.NewSearchProductsUsecase(productRepo)
	paymentUC := usecase.NewPaymentUsecases(paymentRepo, orderRepo)
	orderUC := usecase.NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, locationRepo)
	businessUC := usecase.NewBusinessProfileUsecases(businessRepo)
	clientUC := usecase.NewClientUsecases(clientRepo)
	expenseUC := usecase.NewExpenseUsecases(expenseRepo)
	metricsRepo := postgres.NewPostgresMetricsRepository(db)
	metricsUC := usecase.NewMetricsUsecases(metricsRepo)
	locationUC := usecase.NewLocationUsecases(locationRepo)
	integrationsUC := usecase.NewIntegrationsUsecases(metaRepo)

	// ── Infrastructure: HTTP ─────────────────────────────────────

	authHandler := handler.NewAuthHandler(registerUC, loginUC, resetPasswordUC, userRepo)
	productHandler := handler.NewProductHandler(catalogUC)
	insumoHandler := handler.NewInsumoHandler(insumoUC)
	serviceHandler := handler.NewServiceHandler(serviceUC)
	paymentHandler := handler.NewPaymentHandler(paymentUC)
	orderHandler := handler.NewOrderHandler(orderUC, locationUC)
	metaHandler := handler.NewMetaHandler(metaRepo)
	businessHandler := handler.NewBusinessHandler(businessUC)
	syncJobHandler := handler.NewSyncJobHandler(syncJobRepo)
	toolHandler := handler.NewToolHandler(catalogUC, searchProductsUC, serviceUC)
	businessToolHandler := handler.NewBusinessToolHandler(orderUC, metricsUC, clientUC, insumoUC)
	tagHandler := handler.NewTagHandler(tagUC)
	clientHandler := handler.NewClientHandler(clientUC)
	expenseHandler := handler.NewExpenseHandler(expenseUC)
	metricsHandler := handler.NewMetricsHandler(metricsUC)
	locationHandler := handler.NewLocationHandler(locationUC)
	integrationsHandler := handler.NewIntegrationsHandler(integrationsUC)

	router := http.NewRouter(authHandler, productHandler, insumoHandler, serviceHandler, paymentHandler, orderHandler, metaHandler, businessHandler, syncJobHandler, toolHandler, businessToolHandler, tagHandler, jwtService, cfg.CoreInternalToken, clientHandler, expenseHandler, metricsHandler, locationHandler, integrationsHandler)

	// ── Start Workers ────────────────────────────────────────────
	// Create context that will be cancelled on exit (Run will block, but we can manage context if needed)
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	syncWorker := workers.NewSyncWorker(syncJobRepo)
	go syncWorker.Start(workerCtx)

	// ── Start Server ─────────────────────────────────────────────
	addr := ":" + cfg.ServerPort
	logger.L().Info("🚀 Vento Core server running", "addr", "http://localhost"+addr)
	if err := router.Run(addr); err != nil {
		logger.L().Error("❌ Server failed", "error", err)
		os.Exit(1)
	}
}
