package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
)

// fakeMetricsRepo satisfies port.MetricsRepository for handler tests.
type fakeMetricsRepo struct {
	fetchMetricsFunc func(ctx context.Context, userID string, days int) (*dto.MetricsResponse, error)
}

func (f *fakeMetricsRepo) FetchMetrics(ctx context.Context, userID string, days int) (*dto.MetricsResponse, error) {
	if f.fetchMetricsFunc != nil {
		return f.fetchMetricsFunc(ctx, userID, days)
	}
	return &dto.MetricsResponse{}, nil
}

// fakeClientStatsRepo satisfies port.ClientRepository for handler tests.
type fakeClientStatsRepo struct {
	listFunc func(ctx context.Context, userID string) ([]port.ClientStats, error)
}

func (f *fakeClientStatsRepo) Save(ctx context.Context, client *entity.Client) error { return nil }
func (f *fakeClientStatsRepo) ListByUserIDWithStats(ctx context.Context, userID string) ([]port.ClientStats, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, userID)
	}
	return nil, nil
}

func newOrderFor(t *testing.T, status entity.OrderStatus, total float64, createdAt time.Time) *entity.Order {
	t.Helper()
	o := entity.NewOrder("user-123", "Cliente Test", "", []entity.OrderItem{
		{ProductID: "p1", Name: "Item", Quantity: 1, UnitPrice: total},
	})
	o.Status = status
	o.Total = total
	o.CreatedAt = createdAt
	o.UpdatedAt = createdAt
	return o
}

func TestBusinessToolHandler_GetOrders(t *testing.T) {
	t.Run("scopes to tenant and returns counts_by_status", func(t *testing.T) {
		now := time.Now()
		older := now.Add(-1 * time.Hour)
		repo := &fakeOrderRepo{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]*entity.Order, error) {
				assert.Equal(t, "user-123", userID)
				return []*entity.Order{
					newOrderFor(t, entity.StatusPending, 100, older),
					newOrderFor(t, entity.StatusConfirmed, 200, now),
				}, nil
			},
		}
		orderUC := usecase.NewOrderUsecases(repo, &fakeServiceRepo{}, &fakeProductRepo{}, &fakeInsumoRepo{}, &fakeLocationRepo{})
		h := handler.NewBusinessToolHandler(orderUC, nil, nil, nil)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/orders", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetOrders(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/orders", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body struct {
			Orders         []dto.OrderResponse `json:"orders"`
			CountsByStatus map[string]int       `json:"counts_by_status"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body.Orders, 2)
		// most recent first
		assert.Equal(t, "confirmed", body.Orders[0].Status)
		assert.Equal(t, 1, body.CountsByStatus["pending"])
		assert.Equal(t, 1, body.CountsByStatus["confirmed"])
	})

	t.Run("caps orders to 20 most recent", func(t *testing.T) {
		orders := make([]*entity.Order, 0, 25)
		base := time.Now()
		for i := 0; i < 25; i++ {
			orders = append(orders, newOrderFor(t, entity.StatusPending, 10, base.Add(time.Duration(i)*time.Minute)))
		}
		repo := &fakeOrderRepo{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]*entity.Order, error) {
				return orders, nil
			},
		}
		orderUC := usecase.NewOrderUsecases(repo, &fakeServiceRepo{}, &fakeProductRepo{}, &fakeInsumoRepo{}, &fakeLocationRepo{})
		h := handler.NewBusinessToolHandler(orderUC, nil, nil, nil)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/orders", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetOrders(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/orders", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body struct {
			Orders []dto.OrderResponse `json:"orders"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body.Orders, 20)
		// newest first: the last created order (index 24) must be first
		assert.Equal(t, orders[24].ID, body.Orders[0].ID)
	})
}

func TestBusinessToolHandler_GetMetrics(t *testing.T) {
	t.Run("defaults to period 7 and passes through response", func(t *testing.T) {
		repo := &fakeMetricsRepo{
			fetchMetricsFunc: func(ctx context.Context, userID string, days int) (*dto.MetricsResponse, error) {
				assert.Equal(t, "user-123", userID)
				assert.Equal(t, 7, days)
				return &dto.MetricsResponse{Summary: dto.MetricsSummary{TotalRevenue: 500}}, nil
			},
		}
		metricsUC := usecase.NewMetricsUsecases(repo)
		h := handler.NewBusinessToolHandler(nil, metricsUC, nil, nil)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/metrics", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetMetrics(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/metrics", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body dto.MetricsResponse
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, float64(500), body.Summary.TotalRevenue)
	})

	t.Run("honors period=1 for revenue today", func(t *testing.T) {
		repo := &fakeMetricsRepo{
			fetchMetricsFunc: func(ctx context.Context, userID string, days int) (*dto.MetricsResponse, error) {
				assert.Equal(t, 1, days)
				return &dto.MetricsResponse{}, nil
			},
		}
		metricsUC := usecase.NewMetricsUsecases(repo)
		h := handler.NewBusinessToolHandler(nil, metricsUC, nil, nil)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/metrics", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetMetrics(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/metrics?period=1", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestBusinessToolHandler_GetClients(t *testing.T) {
	makeStats := func() []port.ClientStats {
		activeDate := time.Now().Add(-5 * 24 * time.Hour)
		riskDate := time.Now().Add(-45 * 24 * time.Hour)
		return []port.ClientStats{
			{
				Client:           &entity.Client{ID: "c1", Name: "Ana"},
				TotalSpent:       1000,
				OrderCount:       3,
				LastPurchaseDate: &activeDate,
			},
			{
				Client:           &entity.Client{ID: "c2", Name: "Beto"},
				TotalSpent:       5000,
				OrderCount:       1,
				LastPurchaseDate: &riskDate,
			},
			{
				Client:     &entity.Client{ID: "c3", Name: "Caro"},
				TotalSpent: 0,
				OrderCount: 0,
			},
		}
	}

	t.Run("returns best client and counts by segment", func(t *testing.T) {
		repo := &fakeClientStatsRepo{
			listFunc: func(ctx context.Context, userID string) ([]port.ClientStats, error) {
				assert.Equal(t, "user-123", userID)
				return makeStats(), nil
			},
		}
		clientUC := usecase.NewClientUsecases(repo)
		h := handler.NewBusinessToolHandler(nil, nil, clientUC, nil)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/clients", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetClients(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/clients", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body struct {
			Clients    []dto.ClientResponse `json:"clients"`
			BestClient *struct {
				Name       string  `json:"name"`
				TotalSpent float64 `json:"total_spent"`
			} `json:"best_client"`
			Counts map[string]int `json:"counts"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body.Clients, 3)
		assert.NotNil(t, body.BestClient)
		assert.Equal(t, "Beto", body.BestClient.Name)
		assert.Equal(t, float64(5000), body.BestClient.TotalSpent)
		assert.Equal(t, 1, body.Counts["active"])
		assert.Equal(t, 1, body.Counts["at_risk"])
		assert.Equal(t, 1, body.Counts["inactive"])
	})

	t.Run("filters by segment", func(t *testing.T) {
		repo := &fakeClientStatsRepo{
			listFunc: func(ctx context.Context, userID string) ([]port.ClientStats, error) {
				return makeStats(), nil
			},
		}
		clientUC := usecase.NewClientUsecases(repo)
		h := handler.NewBusinessToolHandler(nil, nil, clientUC, nil)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/clients", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetClients(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/clients?segment=at_risk", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body struct {
			Clients []dto.ClientResponse `json:"clients"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body.Clients, 1)
		assert.Equal(t, "Beto", body.Clients[0].Name)
	})

	t.Run("empty list returns nil best client", func(t *testing.T) {
		repo := &fakeClientStatsRepo{
			listFunc: func(ctx context.Context, userID string) ([]port.ClientStats, error) {
				return []port.ClientStats{}, nil
			},
		}
		clientUC := usecase.NewClientUsecases(repo)
		h := handler.NewBusinessToolHandler(nil, nil, clientUC, nil)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/clients", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetClients(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/clients", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body struct {
			Clients    []dto.ClientResponse `json:"clients"`
			BestClient *struct{}            `json:"best_client"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body.Clients, 0)
		assert.Nil(t, body.BestClient)
	})
}

func TestBusinessToolHandler_GetInsumos(t *testing.T) {
	makeInsumos := func() []*entity.Insumo {
		lowStock := 3.0
		highStock := 50.0
		return []*entity.Insumo{
			{ID: "i1", Name: "Harina 000", Category: "seco", Stock: &lowStock, StockUnit: "kg"},
			{ID: "i2", Name: "Azúcar", Category: "seco", Stock: &highStock, StockUnit: "kg"},
			{ID: "i3", Name: "Manteca", Category: "lacteo", Stock: nil, StockUnit: "kg"},
		}
	}

	t.Run("returns tenant-scoped insumos", func(t *testing.T) {
		insumoRepo := &fakeInsumoRepo{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]*entity.Insumo, error) {
				assert.Equal(t, "user-123", userID)
				return makeInsumos(), nil
			},
		}
		insumoUC := usecase.NewInsumoUsecases(insumoRepo, &fakeTagRepo{}, &fakeLocationRepo{})
		h := handler.NewBusinessToolHandler(nil, nil, nil, insumoUC)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/insumos", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetInsumos(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/insumos", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body struct {
			Insumos []dto.InsumoResponse `json:"insumos"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body.Insumos, 3)
	})

	t.Run("filters by name query, case/accent-insensitive", func(t *testing.T) {
		insumoRepo := &fakeInsumoRepo{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]*entity.Insumo, error) {
				return makeInsumos(), nil
			},
		}
		insumoUC := usecase.NewInsumoUsecases(insumoRepo, &fakeTagRepo{}, &fakeLocationRepo{})
		h := handler.NewBusinessToolHandler(nil, nil, nil, insumoUC)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/insumos", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetInsumos(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/insumos?q=azucar", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body struct {
			Insumos []dto.InsumoResponse `json:"insumos"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body.Insumos, 1)
		assert.Equal(t, "Azúcar", body.Insumos[0].Name)
	})

	t.Run("filters by low_stock threshold", func(t *testing.T) {
		insumoRepo := &fakeInsumoRepo{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]*entity.Insumo, error) {
				return makeInsumos(), nil
			},
		}
		insumoUC := usecase.NewInsumoUsecases(insumoRepo, &fakeTagRepo{}, &fakeLocationRepo{})
		h := handler.NewBusinessToolHandler(nil, nil, nil, insumoUC)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/insumos", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetInsumos(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/insumos?low_stock=true", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body struct {
			Insumos []dto.InsumoResponse `json:"insumos"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body.Insumos, 1)
		assert.Equal(t, "Harina 000", body.Insumos[0].Name)
	})

	t.Run("empty list returns empty array not null", func(t *testing.T) {
		insumoRepo := &fakeInsumoRepo{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]*entity.Insumo, error) {
				return []*entity.Insumo{}, nil
			},
		}
		insumoUC := usecase.NewInsumoUsecases(insumoRepo, &fakeTagRepo{}, &fakeLocationRepo{})
		h := handler.NewBusinessToolHandler(nil, nil, nil, insumoUC)

		r, v1 := setupTestRouter()
		v1.GET("/internal/tools/insumos", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.GetInsumos(c)
		})

		w := performRequest(r, "GET", "/api/v1/internal/tools/insumos", nil)

		assert.Equal(t, http.StatusOK, w.Code)
		var body struct {
			Insumos []dto.InsumoResponse `json:"insumos"`
		}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Len(t, body.Insumos, 0)
	})
}
