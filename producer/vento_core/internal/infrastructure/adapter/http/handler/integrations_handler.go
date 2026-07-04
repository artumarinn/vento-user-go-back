package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type IntegrationsHandler struct {
	integrationsUC *usecase.IntegrationsUsecases
}

func NewIntegrationsHandler(integrationsUC *usecase.IntegrationsUsecases) *IntegrationsHandler {
	return &IntegrationsHandler{integrationsUC: integrationsUC}
}

// GetStatus reports which channels are really connected for the authenticated user.
// GET /api/v1/integrations/status
func (h *IntegrationsHandler) GetStatus(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	status, err := h.integrationsUC.GetStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch integrations status"})
		return
	}

	c.JSON(http.StatusOK, status)
}
