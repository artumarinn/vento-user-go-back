package http

import (
	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/middleware"
)

// NewRouter creates and configures the Gin router with all routes.
func NewRouter(authHandler *handler.AuthHandler, productHandler *handler.ProductHandler, paymentHandler *handler.PaymentHandler, orderHandler *handler.OrderHandler, metaHandler *handler.MetaHandler, businessHandler *handler.BusinessHandler, authService port.AuthService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "vento_core"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Internal routes (No Auth required for now, or use a shared secret)
		internal := v1.Group("/internal")
		{
			internal.GET("/meta-config/:platformID", metaHandler.GetByPlatformID)
			internal.GET("/business-profile/:userID", businessHandler.GetProfileInternal)
			internal.GET("/products/:userID", productHandler.ListInternal)
			internal.POST("/products/sync", productHandler.SyncFromIA)
			internal.POST("/orders/sync", orderHandler.SyncFromIA)
			internal.POST("/payments/sync", paymentHandler.SyncFromIA)
		}
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.GET("/me", middleware.AuthMiddleware(authService), authHandler.Me)
		}

		products := v1.Group("/products")
		products.Use(middleware.AuthMiddleware(authService))
		{
			products.GET("", productHandler.List)
			products.POST("", productHandler.Create)
			products.POST("/batch", productHandler.CreateBatch)
			products.PUT("/:id", productHandler.Update)
			products.DELETE("/:id", productHandler.Delete)
		}

		// Alias for Frontend compatibility
		catalog := v1.Group("/catalog")
		catalog.Use(middleware.AuthMiddleware(authService))
		{
			catalog.GET("", productHandler.List)
			catalog.POST("", productHandler.Create)
		}

		business := v1.Group("/business-profile")
		business.Use(middleware.AuthMiddleware(authService))
		{
			business.GET("", businessHandler.GetProfile)
			business.POST("", businessHandler.SaveProfile)
		}

		dashboard := v1.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware(authService))
		{
			dashboard.GET("/stats", businessHandler.GetStats)
		}

		payments := v1.Group("/payments")
		payments.Use(middleware.AuthMiddleware(authService))
		{
			payments.GET("", paymentHandler.List)
		}
	}

	return r
}
