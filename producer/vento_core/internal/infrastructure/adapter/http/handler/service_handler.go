package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type ServiceHandler struct {
	serviceUC *usecase.ServiceUsecases
}

func NewServiceHandler(serviceUC *usecase.ServiceUsecases) *ServiceHandler {
	return &ServiceHandler{serviceUC: serviceUC}
}

func (h *ServiceHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	services, err := h.serviceUC.ListServices(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, services)
}

func (h *ServiceHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := h.serviceUC.CreateService(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, service)
}

func (h *ServiceHandler) CreateBatch(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.BatchCreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	services, err := h.serviceUC.BatchCreateServices(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, services)
}

func (h *ServiceHandler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	serviceID := c.Param("id")
	var req dto.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := h.serviceUC.UpdateService(c.Request.Context(), userID, serviceID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, service)
}

func (h *ServiceHandler) SyncFromIA(c *gin.Context) {
	logger.L().Debug("Core: Received SyncFromIA request")
	var req dto.BatchCreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Error("Core: Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Services) == 0 {
		logger.L().Warn("Core: No services received in sync request")
		c.JSON(http.StatusOK, gin.H{"message": "no services to sync"})
		return
	}

	tenantID, ok := c.Get("tenantUserID")
	userID, _ := tenantID.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenant context"})
		return
	}
	logger.L().Debug("Core: Syncing services from IA", "userID", userID, "count", len(req.Services))

	services, err := h.serviceUC.BatchCreateServices(c.Request.Context(), userID, req)
	if err != nil {
		logger.L().Error("Core: BatchCreateServices failed", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.L().Info("Core: Successfully synced services", "userID", userID, "count", len(services))
	c.JSON(http.StatusCreated, services)
}

// PricePreview computes the price a Service's formula would yield for the
// submitted variable values, reusing the same evaluator path used at order
// creation time (single source of truth — no client-side formula mirror).
func (h *ServiceHandler) PricePreview(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	serviceID := c.Param("id")

	var req dto.PreviewPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unitPrice, err := h.serviceUC.PreviewPrice(c.Request.Context(), userID, serviceID, req.Variables)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.PreviewPriceResponse{
		UnitPrice: unitPrice,
		LineTotal: unitPrice,
	})
}

// PreviewPriceDraft computes the price for an in-progress service definition
// (formula + variables_schema not yet saved), used by the catalog "probá tu
// precio" sandbox so a service author can validate pricing before saving.
func (h *ServiceHandler) PreviewPriceDraft(c *gin.Context) {
	var req dto.PreviewPriceDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unitPrice, err := h.serviceUC.PreviewPriceDraft(req.Formula, req.VariablesSchema, req.Variables)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.PreviewPriceResponse{
		UnitPrice: unitPrice,
		LineTotal: unitPrice,
	})
}

func (h *ServiceHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	serviceID := c.Param("id")

	if err := h.serviceUC.DeleteService(c.Request.Context(), userID, serviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
