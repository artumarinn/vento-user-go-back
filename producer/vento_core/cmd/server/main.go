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

	// ── Application: Use Cases ───────────────────────────────────
	registerUC := usecase.NewRegisterUser(userRepo, jwtService)
	loginUC := usecase.NewLoginUser(userRepo, jwtService)
	resetPasswordUC := usecase.NewResetPasswordUsecases(userRepo)
	catalogUC := usecase.NewCatalogUsecases(productRepo, tagRepo)
	insumoUC := usecase.NewInsumoUsecases(insumoRepo, tagRepo)
	serviceUC := usecase.NewServiceUsecases(serviceRepo, tagRepo)
	tagUC := usecase.NewTagUsecases(tagRepo)
	searchProductsUC := usecase.NewSearchProductsUsecase(productRepo)
	paymentUC := usecase.NewPaymentUsecases(paymentRepo, orderRepo)
	orderUC := usecase.NewOrderUsecases(orderRepo, serviceRepo)
	businessUC := usecase.NewBusinessProfileUsecases(businessRepo)
	clientUC := usecase.NewClientUsecases(clientRepo)

	// ── Infrastructure: HTTP ─────────────────────────────────────
	authHandler := handler.NewAuthHandler(registerUC, loginUC, resetPasswordUC, userRepo)
	productHandler := handler.NewProductHandler(catalogUC)
	insumoHandler := handler.NewInsumoHandler(insumoUC)
	serviceHandler := handler.NewServiceHandler(serviceUC)
	paymentHandler := handler.NewPaymentHandler(paymentUC)
	orderHandler := handler.NewOrderHandler(orderUC)
	metaHandler := handler.NewMetaHandler(metaRepo)
	businessHandler := handler.NewBusinessHandler(businessUC)
	syncJobHandler := handler.NewSyncJobHandler(syncJobRepo)
	toolHandler := handler.NewToolHandler(catalogUC, searchProductsUC)
	tagHandler := handler.NewTagHandler(tagUC)
	clientHandler := handler.NewClientHandler(clientUC)

	router := http.NewRouter(authHandler, productHandler, insumoHandler, serviceHandler, paymentHandler, orderHandler, metaHandler, businessHandler, syncJobHandler, toolHandler, tagHandler, jwtService, cfg.CoreInternalToken, clientHandler)

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
