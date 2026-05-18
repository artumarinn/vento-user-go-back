package main

import (
	"os"

	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/postgres"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/auth"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/config"
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
	paymentRepo := postgres.NewPostgresPaymentRepository(db)
	orderRepo := postgres.NewPostgresOrderRepository(db)
	metaRepo := postgres.NewMetaRepo(db)
	businessRepo := postgres.NewBusinessRepo(db)
	syncJobRepo := postgres.NewPostgresSyncJobRepository(db)

	// ── Application: Use Cases ───────────────────────────────────
	registerUC := usecase.NewRegisterUser(userRepo, jwtService)
	loginUC := usecase.NewLoginUser(userRepo, jwtService)
	resetPasswordUC := usecase.NewResetPasswordUsecases(userRepo)
	catalogUC := usecase.NewCatalogUsecases(productRepo)
	paymentUC := usecase.NewPaymentUsecases(paymentRepo)
	orderUC := usecase.NewOrderUsecases(orderRepo)
	businessUC := usecase.NewBusinessProfileUsecases(businessRepo)

	// ── Infrastructure: HTTP ─────────────────────────────────────
	authHandler := handler.NewAuthHandler(registerUC, loginUC, resetPasswordUC, userRepo)
	productHandler := handler.NewProductHandler(catalogUC)
	paymentHandler := handler.NewPaymentHandler(paymentUC)
	orderHandler := handler.NewOrderHandler(orderUC)
	metaHandler := handler.NewMetaHandler(metaRepo)
	businessHandler := handler.NewBusinessHandler(businessUC)
	syncJobHandler := handler.NewSyncJobHandler(syncJobRepo)

	router := http.NewRouter(authHandler, productHandler, paymentHandler, orderHandler, metaHandler, businessHandler, syncJobHandler, jwtService, cfg.CoreInternalToken)


	// ── Start Server ─────────────────────────────────────────────
	addr := ":" + cfg.ServerPort
	logger.L().Info("🚀 Vento Core server running", "addr", "http://localhost"+addr)
	if err := router.Run(addr); err != nil {
		logger.L().Error("❌ Server failed", "error", err)
		os.Exit(1)
	}
}
