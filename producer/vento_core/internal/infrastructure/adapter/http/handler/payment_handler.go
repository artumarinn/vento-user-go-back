package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type PaymentHandler struct {
	paymentUC *usecase.PaymentUsecases
}

func NewPaymentHandler(paymentUC *usecase.PaymentUsecases) *PaymentHandler {
	return &PaymentHandler{paymentUC: paymentUC}
}

func (h *PaymentHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	payments, err := h.paymentUC.ListPayments(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payments)
}
