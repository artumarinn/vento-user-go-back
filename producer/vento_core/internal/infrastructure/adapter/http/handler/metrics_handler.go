package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type MetricsHandler struct {
	metricsUC *usecase.MetricsUsecases
}

func NewMetricsHandler(metricsUC *usecase.MetricsUsecases) *MetricsHandler {
	return &MetricsHandler{metricsUC: metricsUC}
}

func (h *MetricsHandler) Get(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	period := c.DefaultQuery("period", "7d")
	data, err := h.metricsUC.GetMetrics(c.Request.Context(), userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
