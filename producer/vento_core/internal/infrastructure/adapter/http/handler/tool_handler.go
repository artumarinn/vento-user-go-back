package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

// ServiceToolUsecases is the narrow surface ToolHandler needs from
// ServiceUsecases — lets tests inject a fake without depending on the
// concrete usecase struct.
type ServiceToolUsecases interface {
	SearchServices(ctx context.Context, userID, query string) ([]dto.ServiceSearchResult, error)
	CheckFeasibility(ctx context.Context, userID, serviceID string, quantity float64) (offered bool, feasible bool, err error)
	GetServiceVariables(ctx context.Context, userID, serviceID string) ([]dto.ServiceVariableResponse, error)
	GetServicePrice(ctx context.Context, userID, serviceID string, variables map[string]any) (float64, error)
}

type ToolHandler struct {
	catalogUC *usecase.CatalogUsecases
	searchUC  *usecase.SearchProductsUsecase
	serviceUC ServiceToolUsecases
}

func NewToolHandler(catalogUC *usecase.CatalogUsecases, searchUC *usecase.SearchProductsUsecase, serviceUC ServiceToolUsecases) *ToolHandler {
	return &ToolHandler{
		catalogUC: catalogUC,
		searchUC:  searchUC,
		serviceUC: serviceUC,
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

func (h *ToolHandler) SearchServices(c *gin.Context) {
	tenantID, _ := c.Get("tenantUserID")
	userID := tenantID.(string)
	q := strings.TrimSpace(c.Query("q"))

	results, err := h.serviceUC.SearchServices(c.Request.Context(), userID, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

type checkFeasibilityInput struct {
	ServiceID string  `json:"service_id"`
	Quantity  float64 `json:"quantity"`
}

func (h *ToolHandler) CheckServiceFeasibility(c *gin.Context) {
	tenantID, _ := c.Get("tenantUserID")
	userID := tenantID.(string)

	var input checkFeasibilityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offered, feasible, err := h.serviceUC.CheckFeasibility(c.Request.Context(), userID, input.ServiceID, input.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"offered": offered, "feasible": feasible})
}

func (h *ToolHandler) GetServiceVariables(c *gin.Context) {
	tenantID, _ := c.Get("tenantUserID")
	userID := tenantID.(string)
	serviceID := c.Query("service_id")

	variables, err := h.serviceUC.GetServiceVariables(c.Request.Context(), userID, serviceID)
	if err != nil {
		if errors.Is(err, usecase.ErrServiceNotFoundForPricing) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, variables)
}

type getServicePriceInput struct {
	ServiceID string         `json:"service_id"`
	Variables map[string]any `json:"variables"`
}

func (h *ToolHandler) GetServicePrice(c *gin.Context) {
	tenantID, _ := c.Get("tenantUserID")
	userID := tenantID.(string)

	var input getServicePriceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	price, err := h.serviceUC.GetServicePrice(c.Request.Context(), userID, input.ServiceID, input.Variables)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrServiceNotFoundForPricing):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrMissingRequiredVariable),
			errors.Is(err, usecase.ErrUnknownVariableName),
			errors.Is(err, usecase.ErrInvalidVariableOption),
			errors.Is(err, usecase.ErrInvalidVariableValue):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"price": price})
}
