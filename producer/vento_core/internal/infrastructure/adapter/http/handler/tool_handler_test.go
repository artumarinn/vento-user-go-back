package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
)

type fakeServiceSearchUC struct {
	searchFunc       func(ctx context.Context, userID, query string) ([]dto.ServiceSearchResult, error)
	feasibilityFunc  func(ctx context.Context, userID, serviceID string, quantity float64) (bool, bool, error)
	getVariablesFunc func(ctx context.Context, userID, serviceID string) ([]dto.ServiceVariableResponse, error)
	getPriceFunc     func(ctx context.Context, userID, serviceID string, variables map[string]any) (float64, error)
}

func (f *fakeServiceSearchUC) SearchServices(ctx context.Context, userID, query string) ([]dto.ServiceSearchResult, error) {
	return f.searchFunc(ctx, userID, query)
}

func (f *fakeServiceSearchUC) CheckFeasibility(ctx context.Context, userID, serviceID string, quantity float64) (bool, bool, error) {
	return f.feasibilityFunc(ctx, userID, serviceID, quantity)
}

func (f *fakeServiceSearchUC) GetServiceVariables(ctx context.Context, userID, serviceID string) ([]dto.ServiceVariableResponse, error) {
	return f.getVariablesFunc(ctx, userID, serviceID)
}

func (f *fakeServiceSearchUC) GetServicePrice(ctx context.Context, userID, serviceID string, variables map[string]any) (float64, error) {
	return f.getPriceFunc(ctx, userID, serviceID, variables)
}

func setupToolRouter(t *testing.T, uc handler.ServiceToolUsecases) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewToolHandler(nil, nil, uc)
	r.Use(func(c *gin.Context) {
		c.Set("tenantUserID", "u1")
		c.Next()
	})
	r.GET("/tools/services/search", h.SearchServices)
	r.POST("/tools/services/feasibility", h.CheckServiceFeasibility)
	r.GET("/tools/services/variables", h.GetServiceVariables)
	r.POST("/tools/services/price", h.GetServicePrice)
	return r
}

func TestSearchServices_ReturnsResults(t *testing.T) {
	uc := &fakeServiceSearchUC{
		searchFunc: func(ctx context.Context, userID, query string) ([]dto.ServiceSearchResult, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, "llavero", query)
			return []dto.ServiceSearchResult{{ID: "s1", Name: "Llaveros 3D", MinimumLeadTime: 2}}, nil
		},
	}
	r := setupToolRouter(t, uc)

	req := httptest.NewRequest(http.MethodGet, "/tools/services/search?q=llavero", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var results []dto.ServiceSearchResult
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &results))
	assert.Len(t, results, 1)
	assert.Equal(t, "s1", results[0].ID)
}

func TestCheckServiceFeasibility_ReturnsOfferedAndFeasible(t *testing.T) {
	uc := &fakeServiceSearchUC{
		feasibilityFunc: func(ctx context.Context, userID, serviceID string, quantity float64) (bool, bool, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, "s1", serviceID)
			assert.Equal(t, 50.0, quantity)
			return true, true, nil
		},
	}
	r := setupToolRouter(t, uc)

	body, _ := json.Marshal(map[string]any{"service_id": "s1", "quantity": 50})
	req := httptest.NewRequest(http.MethodPost, "/tools/services/feasibility", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]bool
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["offered"])
	assert.True(t, resp["feasible"])
}

func TestCheckServiceFeasibility_NotOffered(t *testing.T) {
	uc := &fakeServiceSearchUC{
		feasibilityFunc: func(ctx context.Context, userID, serviceID string, quantity float64) (bool, bool, error) {
			return false, false, nil
		},
	}
	r := setupToolRouter(t, uc)

	body, _ := json.Marshal(map[string]any{"service_id": "nope", "quantity": 1})
	req := httptest.NewRequest(http.MethodPost, "/tools/services/feasibility", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]bool
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp["offered"])
	assert.False(t, resp["feasible"])
}

func TestGetServiceVariables_ReturnsCustomerSafeList(t *testing.T) {
	uc := &fakeServiceSearchUC{
		getVariablesFunc: func(ctx context.Context, userID, serviceID string) ([]dto.ServiceVariableResponse, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, "s1", serviceID)
			return []dto.ServiceVariableResponse{
				{Name: "color", Label: "Color", Type: "select", Required: true,
					Options: []dto.ServiceVariableOption{{Label: "Negro", Value: "negro"}}},
			}, nil
		},
	}
	r := setupToolRouter(t, uc)

	req := httptest.NewRequest(http.MethodGet, "/tools/services/variables?service_id=s1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var results []dto.ServiceVariableResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &results))
	assert.Len(t, results, 1)
	assert.Equal(t, "color", results[0].Name)
}

func TestGetServiceVariables_NotFound(t *testing.T) {
	uc := &fakeServiceSearchUC{
		getVariablesFunc: func(ctx context.Context, userID, serviceID string) ([]dto.ServiceVariableResponse, error) {
			return nil, usecase.ErrServiceNotFoundForPricing
		},
	}
	r := setupToolRouter(t, uc)

	req := httptest.NewRequest(http.MethodGet, "/tools/services/variables?service_id=nope", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetServicePrice_ReturnsPrice(t *testing.T) {
	uc := &fakeServiceSearchUC{
		getPriceFunc: func(ctx context.Context, userID, serviceID string, variables map[string]any) (float64, error) {
			assert.Equal(t, "u1", userID)
			assert.Equal(t, "s1", serviceID)
			assert.Equal(t, "negro", variables["color"])
			return 5200.0, nil
		},
	}
	r := setupToolRouter(t, uc)

	body, _ := json.Marshal(map[string]any{"service_id": "s1", "variables": map[string]any{"color": "negro"}})
	req := httptest.NewRequest(http.MethodPost, "/tools/services/price", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]float64
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 5200.0, resp["price"])
}

func TestGetServicePrice_MissingRequiredVariableReturns400(t *testing.T) {
	uc := &fakeServiceSearchUC{
		getPriceFunc: func(ctx context.Context, userID, serviceID string, variables map[string]any) (float64, error) {
			return 0, usecase.ErrMissingRequiredVariable
		},
	}
	r := setupToolRouter(t, uc)

	body, _ := json.Marshal(map[string]any{"service_id": "s1", "variables": map[string]any{}})
	req := httptest.NewRequest(http.MethodPost, "/tools/services/price", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetServicePrice_ServiceNotFoundReturns404(t *testing.T) {
	uc := &fakeServiceSearchUC{
		getPriceFunc: func(ctx context.Context, userID, serviceID string, variables map[string]any) (float64, error) {
			return 0, usecase.ErrServiceNotFoundForPricing
		},
	}
	r := setupToolRouter(t, uc)

	body, _ := json.Marshal(map[string]any{"service_id": "nope", "variables": map[string]any{}})
	req := httptest.NewRequest(http.MethodPost, "/tools/services/price", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
