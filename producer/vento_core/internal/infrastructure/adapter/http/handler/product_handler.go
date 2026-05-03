package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type ProductHandler struct {
	catalogUC *usecase.CatalogUsecases
}

func NewProductHandler(catalogUC *usecase.CatalogUsecases) *ProductHandler {
	return &ProductHandler{catalogUC: catalogUC}
}

func (h *ProductHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	products, err := h.catalogUC.ListProducts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) ListInternal(c *gin.Context) {
	userID := c.Param("userID")
	logger.L().Debug("Core: Internal request to list products", "userID", userID)
	products, err := h.catalogUC.ListProducts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.catalogUC.CreateProduct(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) CreateBatch(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.BatchCreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	products, err := h.catalogUC.BatchCreateProducts(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, products)
}

func (h *ProductHandler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	productID := c.Param("id")
	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.catalogUC.UpdateProduct(c.Request.Context(), userID, productID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) SyncFromIA(c *gin.Context) {
	logger.L().Debug("Core: Received SyncFromIA request")
	var req dto.BatchCreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.L().Error("Core: Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Products) == 0 {
		logger.L().Warn("Core: No products received in sync request")
		c.JSON(http.StatusOK, gin.H{"message": "no products to sync"})
		return
	}

	userID := req.Products[0].UserID 
	logger.L().Debug("Core: Syncing products from IA", "userID", userID, "count", len(req.Products))

	products, err := h.catalogUC.BatchCreateProducts(c.Request.Context(), userID, req)
	if err != nil {
		logger.L().Error("Core: BatchCreateProducts failed", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.L().Info("Core: Successfully synced products", "userID", userID, "count", len(products))
	c.JSON(http.StatusCreated, products)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	productID := c.Param("id")

	if err := h.catalogUC.DeleteProduct(c.Request.Context(), userID, productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
