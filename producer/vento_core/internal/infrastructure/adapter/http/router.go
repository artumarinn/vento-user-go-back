package http

import (
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/middleware"
)

// NewRouter creates and configures the Gin router with all routes.
func NewRouter(authHandler *handler.AuthHandler, productHandler *handler.ProductHandler, insumoHandler *handler.InsumoHandler, serviceHandler *handler.ServiceHandler, paymentHandler *handler.PaymentHandler, orderHandler *handler.OrderHandler, metaHandler *handler.MetaHandler, businessHandler *handler.BusinessHandler, syncJobHandler *handler.SyncJobHandler, toolHandler *handler.ToolHandler, businessToolHandler *handler.BusinessToolHandler, tagHandler *handler.TagHandler, authService port.AuthService, internalToken string, clientHandler *handler.ClientHandler, expenseHandler *handler.ExpenseHandler, metricsHandler *handler.MetricsHandler, locationHandler *handler.LocationHandler, integrationsHandler *handler.IntegrationsHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Use custom slog-based logger
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			if param.Path == "/health" {
				return ""
			}
			return fmt.Sprintf("[GIN] %s | %d | %s | %s | %s %s | %s\n",
				param.TimeStamp.Format(time.RFC3339),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
				param.ErrorMessage,
			)
		},
		Output: io.Discard, // We log via slog instead
	}))

	// Log via slog for better integration
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		if path == "/health" {
			return
		}

		end := time.Now()
		latency := end.Sub(start)

		logger.L().Info("http request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"ip", c.ClientIP(),
			"latency", latency,
			"user_agent", c.Request.UserAgent(),
		)
	})

	r.Use(gin.Recovery())

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
		internal := v1.Group("/internal")
		internal.Use(middleware.InternalTokenMiddleware(internalToken))
		{
			internal.GET("/meta-config/:platformID", metaHandler.GetByPlatformID)
			internal.GET("/meta-config-by-user/:userID", metaHandler.GetByUserAndChannel)
			internal.GET("/business-profile/:userID", businessHandler.GetProfileInternal)
			internal.GET("/products/:userID", productHandler.ListInternal)
			internal.POST("/products/sync", productHandler.SyncFromIA)
			internal.POST("/insumos/sync", insumoHandler.SyncFromIA)
			internal.POST("/services/sync", serviceHandler.SyncFromIA)
			internal.POST("/orders/sync", orderHandler.SyncFromIA)
			internal.POST("/payments/sync", paymentHandler.SyncFromIA)

			// Tool Calling Endpoints
			internal.GET("/tools/products/:productId/stock", toolHandler.GetStock)
			internal.GET("/tools/products/:productId", toolHandler.GetProductDetails)
			internal.POST("/tools/products/search", toolHandler.SearchProducts)
			internal.GET("/tools/orders", businessToolHandler.GetOrders)
			internal.GET("/tools/metrics", businessToolHandler.GetMetrics)
			internal.GET("/tools/clients", businessToolHandler.GetClients)
			internal.GET("/tools/insumos", businessToolHandler.GetInsumos)

			// Sync Jobs
			internal.POST("/sync-jobs", syncJobHandler.CreateJob)
			internal.GET("/sync-jobs/:id", syncJobHandler.GetJob)
			internal.PUT("/sync-jobs/:id/progress", syncJobHandler.UpdateProgress)
		}
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/forgot-password", authHandler.ForgotPassword)
			authGroup.POST("/reset-password", authHandler.ResetPassword)
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

		insumos := v1.Group("/insumos")
		insumos.Use(middleware.AuthMiddleware(authService))
		{
			insumos.GET("", insumoHandler.List)
			insumos.POST("", insumoHandler.Create)
			insumos.POST("/batch", insumoHandler.CreateBatch)
			insumos.PUT("/:id", insumoHandler.Update)
			insumos.DELETE("/:id", insumoHandler.Delete)
		}

		services := v1.Group("/services")
		services.Use(middleware.AuthMiddleware(authService))
		{
			services.GET("", serviceHandler.List)
			services.POST("", serviceHandler.Create)
			services.POST("/batch", serviceHandler.CreateBatch)
			services.POST("/preview-price", serviceHandler.PreviewPriceDraft)
			services.PUT("/:id", serviceHandler.Update)
			services.DELETE("/:id", serviceHandler.Delete)
			services.POST("/:id/price-preview", serviceHandler.PricePreview)
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

		orders := v1.Group("/orders")
		orders.Use(middleware.AuthMiddleware(authService))
		{
			orders.GET("", orderHandler.List)
			orders.POST("", orderHandler.Create)
			orders.GET("/:id", orderHandler.GetDetail)
			orders.PATCH("/:id/status", orderHandler.UpdateStatus)
			orders.POST("/:id/payment", orderHandler.RegisterPayment)
			orders.PUT("/:id", orderHandler.Update)
			orders.DELETE("/:id", orderHandler.Delete)
		}

		clients := v1.Group("/clients")
		clients.Use(middleware.AuthMiddleware(authService))
		{
			clients.GET("", clientHandler.List)
			clients.POST("", clientHandler.Create)
		}

		metaConfigs := v1.Group("/meta-configs")
		metaConfigs.Use(middleware.AuthMiddleware(authService))
		{
			metaConfigs.POST("", metaHandler.Save)
			metaConfigs.GET("", metaHandler.ListMine)
		}

		tagGroup := v1.Group("/tags")
		tagGroup.Use(middleware.AuthMiddleware(authService))
		{
			tagGroup.GET("", tagHandler.List)
			tagGroup.POST("", tagHandler.Create)
			tagGroup.DELETE("/:id", tagHandler.Delete)
		}

		expenseGroup := v1.Group("/expenses")
		expenseGroup.Use(middleware.AuthMiddleware(authService))
		{
			expenseGroup.GET("", expenseHandler.List)
			expenseGroup.POST("", expenseHandler.Create)
			expenseGroup.DELETE("/:id", expenseHandler.Delete)
		}

		metricsGroup := v1.Group("/metrics")
		metricsGroup.Use(middleware.AuthMiddleware(authService))
		{
			metricsGroup.GET("", metricsHandler.Get)
		}

		locationGroup := v1.Group("/locations")
		locationGroup.Use(middleware.AuthMiddleware(authService))
		{
			locationGroup.GET("", locationHandler.List)
			locationGroup.POST("", locationHandler.Create)
			locationGroup.PUT("/:id", locationHandler.Update)
			locationGroup.DELETE("/:id", locationHandler.Delete)
		}

		integrationsGroup := v1.Group("/integrations")
		integrationsGroup.Use(middleware.AuthMiddleware(authService))
		{
			integrationsGroup.GET("/status", integrationsHandler.GetStatus)
		}
	}

	return r
}
