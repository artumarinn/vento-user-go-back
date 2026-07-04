package handler

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

// maxToolOrders caps the number of orders returned by GetOrders so the AI
// assistant gets a compact, recency-ordered snapshot instead of the full
// history.
const maxToolOrders = 20

// lowStockThreshold is the simple quantity cutoff used by GetInsumos'
// low_stock filter. Documented here rather than configurable per-tenant.
const lowStockThreshold = 5.0

// BusinessToolHandler exposes read-only internal tool endpoints (orders,
// metrics, clients, insumos) consumed by the AI assistant's tool calling.
// Mirrors ToolHandler's pattern: tenant scoping via tenantUserID set by
// InternalTokenMiddleware, read-only reuse of existing usecases.
type BusinessToolHandler struct {
	orderUC   *usecase.OrderUsecases
	metricsUC *usecase.MetricsUsecases
	clientUC  *usecase.ClientUsecases
	insumoUC  *usecase.InsumoUsecases
}

func NewBusinessToolHandler(orderUC *usecase.OrderUsecases, metricsUC *usecase.MetricsUsecases, clientUC *usecase.ClientUsecases, insumoUC *usecase.InsumoUsecases) *BusinessToolHandler {
	return &BusinessToolHandler{
		orderUC:   orderUC,
		metricsUC: metricsUC,
		clientUC:  clientUC,
		insumoUC:  insumoUC,
	}
}

func tenantUserID(c *gin.Context) string {
	tenantID, _ := c.Get("tenantUserID")
	userID, _ := tenantID.(string)
	return userID
}

// GetOrders returns the most recent orders (capped at maxToolOrders) plus a
// tally of orders by status across the full tenant history.
func (h *BusinessToolHandler) GetOrders(c *gin.Context) {
	userID := tenantUserID(c)

	orders, err := h.orderUC.ListOrders(c.Request.Context(), userID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	countsByStatus := make(map[string]int)
	for _, o := range orders {
		countsByStatus[o.Status]++
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].Date.After(orders[j].Date)
	})
	if len(orders) > maxToolOrders {
		orders = orders[:maxToolOrders]
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":           orders,
		"counts_by_status": countsByStatus,
	})
}

// GetMetrics proxies MetricsUsecases.GetMetrics for the requested period
// (defaults to 7 days; period=1 gives "revenue today").
func (h *BusinessToolHandler) GetMetrics(c *gin.Context) {
	userID := tenantUserID(c)
	period := c.DefaultQuery("period", "7")

	metrics, err := h.metricsUC.GetMetrics(c.Request.Context(), userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetClients returns the client list (optionally filtered by segment), the
// client with the highest total spend, and a count per segment.
func (h *BusinessToolHandler) GetClients(c *gin.Context) {
	userID := tenantUserID(c)
	segment := c.Query("segment")

	clients, err := h.clientUC.ListClients(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	counts := map[string]int{"active": 0, "at_risk": 0, "inactive": 0}
	var bestClient *dto.ClientResponse
	for i := range clients {
		if _, tracked := counts[clients[i].Status]; tracked {
			counts[clients[i].Status]++
		}
		if bestClient == nil || clients[i].TotalSpent > bestClient.TotalSpent {
			bestClient = &clients[i]
		}
	}

	filtered := clients
	switch segment {
	case "active", "at_risk", "inactive":
		filtered = make([]dto.ClientResponse, 0, len(clients))
		for _, cl := range clients {
			if cl.Status == segment {
				filtered = append(filtered, cl)
			}
		}
	}

	var bestClientOut interface{}
	if bestClient != nil {
		bestClientOut = gin.H{"name": bestClient.Name, "total_spent": bestClient.TotalSpent}
	}

	c.JSON(http.StatusOK, gin.H{
		"clients":     filtered,
		"best_client": bestClientOut,
		"counts":      counts,
	})
}

// GetInsumos returns the insumo list, optionally filtered by a case/accent
// insensitive name substring (q) and/or a low-stock threshold (low_stock).
func (h *BusinessToolHandler) GetInsumos(c *gin.Context) {
	userID := tenantUserID(c)
	q := strings.TrimSpace(c.Query("q"))
	lowStock, _ := strconv.ParseBool(c.Query("low_stock"))

	insumos, err := h.insumoUC.ListInsumos(c.Request.Context(), userID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filtered := make([]dto.InsumoResponse, 0, len(insumos))
	normalizedQuery := normalizeForSearch(q)
	for _, i := range insumos {
		if normalizedQuery != "" && !strings.Contains(normalizeForSearch(i.Name), normalizedQuery) {
			continue
		}
		if lowStock && (i.Stock == nil || *i.Stock > lowStockThreshold) {
			continue
		}
		filtered = append(filtered, i)
	}

	c.JSON(http.StatusOK, gin.H{"insumos": filtered})
}

// normalizeForSearch lowercases and strips common accented vowels so name
// matching is case/accent-insensitive without pulling in a full Unicode
// normalization dependency.
func normalizeForSearch(s string) string {
	s = strings.ToLower(s)
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	)
	return replacer.Replace(s)
}
