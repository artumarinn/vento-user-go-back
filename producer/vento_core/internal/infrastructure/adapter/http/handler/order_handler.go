package handler

import (
	"errors"
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

// orderErrorStatus maps known domain/pricing errors to their HTTP status,
// falling back to 500 for anything unrecognized.
func orderErrorStatus(err error) int {
	switch {
	case errors.Is(err, usecase.ErrOrderLockedByPayment):
		return http.StatusConflict
	case errors.Is(err, usecase.ErrMissingRequiredVariable),
		errors.Is(err, usecase.ErrInvalidVariableOption),
		errors.Is(err, usecase.ErrUnknownVariableName):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
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

func (h *OrderHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderUC.CreateOrder(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(orderErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetDetail(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	orderID := c.Param("id")

	order, err := h.orderUC.GetOrderDetail(c.Request.Context(), userID, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

type updateOrderStatusRequest struct {
	Status string `json:"status"`
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	orderID := c.Param("id")
	var req updateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderUC.UpdateOrderStatus(c.Request.Context(), userID, orderID, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) RegisterPayment(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	orderID := c.Param("id")
	var req dto.RegisterPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderUC.RegisterPayment(c.Request.Context(), userID, orderID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	orderID := c.Param("id")
	var req dto.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderUC.UpdateOrder(c.Request.Context(), userID, orderID, req)
	if err != nil {
		c.JSON(orderErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	orderID := c.Param("id")

	if err := h.orderUC.DeleteOrder(c.Request.Context(), userID, orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
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
