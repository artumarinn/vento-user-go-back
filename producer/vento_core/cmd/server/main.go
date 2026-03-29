package main

import (
	"log"

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

	// ── Infrastructure: Database ─────────────────────────────────
	db, err := postgres.NewConnection(cfg.DSN())
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()

	// ── Infrastructure: JWT ──────────────────────────────────────
	jwtService := auth.NewJWTService(cfg.JWTSecret)

	// ── Infrastructure: Repositories ─────────────────────────────
	userRepo := postgres.NewPostgresUserRepository(db)
	productRepo := postgres.NewPostgresProductRepository(db)
	paymentRepo := postgres.NewPostgresPaymentRepository(db)

	// ── Application: Use Cases ───────────────────────────────────
	registerUC := usecase.NewRegisterUser(userRepo, jwtService)
	loginUC := usecase.NewLoginUser(userRepo, jwtService)
	catalogUC := usecase.NewCatalogUsecases(productRepo)
	paymentUC := usecase.NewPaymentUsecases(paymentRepo)

	// ── Infrastructure: HTTP ─────────────────────────────────────
	authHandler := handler.NewAuthHandler(registerUC, loginUC, userRepo)
	productHandler := handler.NewProductHandler(catalogUC)
	paymentHandler := handler.NewPaymentHandler(paymentUC)
	router := http.NewRouter(authHandler, productHandler, paymentHandler, jwtService)

	// ── Start Server ─────────────────────────────────────────────
	addr := ":" + cfg.ServerPort
	log.Printf("🚀 Vento Core server running on http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
