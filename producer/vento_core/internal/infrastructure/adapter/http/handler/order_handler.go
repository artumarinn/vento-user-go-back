package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type OrderHandler struct {
	orderUC *usecase.OrderUsecases
}

func NewOrderHandler(orderUC *usecase.OrderUsecases) *OrderHandler {
	return &OrderHandler{orderUC: orderUC}
}

func (h *OrderHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	orders, err := h.orderUC.ListOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) SyncFromIA(c *gin.Context) {
	logger.L().Debug("Core: Received Order Sync request")
	var req dto.BatchCreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Error("Core: Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Orders) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "no orders to sync"})
		return
	}

	userID := req.UserID
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := h.orderUC.SyncFromIA(c.Request.Context(), userID, req)
	if err != nil {
		logger.L().Error("Core: Order Sync failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "orders synced to database"})
}
