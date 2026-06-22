package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type ToolHandler struct {
	catalogUC *usecase.CatalogUsecases
	searchUC  *usecase.SearchProductsUsecase
}

func NewToolHandler(catalogUC *usecase.CatalogUsecases, searchUC *usecase.SearchProductsUsecase) *ToolHandler {
	return &ToolHandler{
		catalogUC: catalogUC,
		searchUC:  searchUC,
	}
}

func (h *ToolHandler) GetStock(c *gin.Context) {
	tenantID, _ := c.Get("tenantUserID")
	userID := tenantID.(string)
	productID := c.Param("productId")

	product, err := h.catalogUC.GetProduct(c.Request.Context(), userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stock":           product.Stock,
		"last_updated_at": product.UpdatedAt,
	})
}

func (h *ToolHandler) GetProductDetails(c *gin.Context) {
	tenantID, _ := c.Get("tenantUserID")
	userID := tenantID.(string)
	productID := c.Param("productId")

	product, err := h.catalogUC.GetProduct(c.Request.Context(), userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ToolHandler) SearchProducts(c *gin.Context) {
	tenantID, _ := c.Get("tenantUserID")
	userID := tenantID.(string)

	var input usecase.SearchProductsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.UserID = userID

	products, err := h.searchUC.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}
