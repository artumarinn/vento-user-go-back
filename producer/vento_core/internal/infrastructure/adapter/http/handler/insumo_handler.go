package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type InsumoHandler struct {
	insumoUC *usecase.InsumoUsecases
}

func NewInsumoHandler(insumoUC *usecase.InsumoUsecases) *InsumoHandler {
	return &InsumoHandler{insumoUC: insumoUC}
}

func (h *InsumoHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	insumos, err := h.insumoUC.ListInsumos(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, insumos)
}

func (h *InsumoHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.CreateInsumoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	insumo, err := h.insumoUC.CreateInsumo(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, insumo)
}

func (h *InsumoHandler) CreateBatch(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.BatchCreateInsumoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	insumos, err := h.insumoUC.BatchCreateInsumos(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, insumos)
}

func (h *InsumoHandler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	insumoID := c.Param("id")
	var req dto.UpdateInsumoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	insumo, err := h.insumoUC.UpdateInsumo(c.Request.Context(), userID, insumoID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, insumo)
}

func (h *InsumoHandler) SyncFromIA(c *gin.Context) {
	logger.L().Debug("Core: Received SyncFromIA request")
	var req dto.BatchCreateInsumoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Error("Core: Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Insumos) == 0 {
		logger.L().Warn("Core: No insumos received in sync request")
		c.JSON(http.StatusOK, gin.H{"message": "no insumos to sync"})
		return
	}

	tenantID, ok := c.Get("tenantUserID")
	userID, _ := tenantID.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenant context"})
		return
	}
	logger.L().Debug("Core: Syncing insumos from IA", "userID", userID, "count", len(req.Insumos))

	insumos, err := h.insumoUC.BatchCreateInsumos(c.Request.Context(), userID, req)
	if err != nil {
		logger.L().Error("Core: BatchCreateInsumos failed", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.L().Info("Core: Successfully synced insumos", "userID", userID, "count", len(insumos))
	c.JSON(http.StatusCreated, insumos)
}

func (h *InsumoHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	insumoID := c.Param("id")

	if err := h.insumoUC.DeleteInsumo(c.Request.Context(), userID, insumoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
