package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
	log.Printf("[DEBUG] Core: Internal request to list products for userID: %s", userID)
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
	log.Printf("[DEBUG] Core: Received SyncFromIA request")
	var req dto.BatchCreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] Core: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] Core: Syncing %d products from IA", len(req.Products))

	if len(req.Products) == 0 {
		log.Printf("[WARN] Core: No products received in sync request")
		c.JSON(http.StatusOK, gin.H{"message": "no products to sync"})
		return
	}

	userID := req.Products[0].UserID 
	log.Printf("[DEBUG] Core: Syncing for userID: %s", userID)

	products, err := h.catalogUC.BatchCreateProducts(c.Request.Context(), userID, req)
	if err != nil {
		log.Printf("[ERROR] Core: BatchCreateProducts failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] Core: Successfully synced %d products", len(products))
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
