package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
)

type fakeProductRepo struct {
	saveFunc func(ctx context.Context, product *entity.Product) error
	updateFunc func(ctx context.Context, product *entity.Product) error
	deleteFunc func(ctx context.Context, id string, userID string) error
	getByIDFunc func(ctx context.Context, id string, userID string) (*entity.Product, error)
	listByUserIDFunc func(ctx context.Context, userID string) ([]*entity.Product, error)
	saveBatchFunc func(ctx context.Context, products []*entity.Product) error
}

func (f *fakeProductRepo) Save(ctx context.Context, p *entity.Product) error { return f.saveFunc(ctx, p) }
func (f *fakeProductRepo) Update(ctx context.Context, p *entity.Product) error { return f.updateFunc(ctx, p) }
func (f *fakeProductRepo) Delete(ctx context.Context, id, u string) error { return f.deleteFunc(ctx, id, u) }
func (f *fakeProductRepo) GetByID(ctx context.Context, id, u string) (*entity.Product, error) { return f.getByIDFunc(ctx, id, u) }
func (f *fakeProductRepo) ListByUserID(ctx context.Context, u string) ([]*entity.Product, error) { return f.listByUserIDFunc(ctx, u) }
func (f *fakeProductRepo) SaveBatch(ctx context.Context, p []*entity.Product) error { return f.saveBatchFunc(ctx, p) }

func TestProductHandler_Create(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		repo := &fakeProductRepo{
			saveFunc: func(ctx context.Context, product *entity.Product) error {
				return nil
			},
		}
		uc := usecase.NewCatalogUsecases(repo)
		h := handler.NewProductHandler(uc)
		r, v1 := setupTestRouter()
		v1.POST("/products", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.Create(c)
		})

		body := gin.H{
			"name":  "Producto Test",
			"sku":   "SKU-123",
			"price": 1500.50,
		}
		w := performRequest(r, "POST", "/api/v1/products", body)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestProductHandler_SyncFromIA(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		repo := &fakeProductRepo{
			saveBatchFunc: func(ctx context.Context, products []*entity.Product) error {
				return nil
			},
		}
		uc := usecase.NewCatalogUsecases(repo)
		h := handler.NewProductHandler(uc)
		r, v1 := setupTestRouter()
		// Inject tenantUserID the same way the internal token middleware does
		v1.POST("/internal/products/sync", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.SyncFromIA(c)
		})

		body := gin.H{
			"products": []gin.H{
				{"user_id": "user-123", "name": "Sync Product", "price": 2000},
			},
		}
		w := performRequest(r, "POST", "/api/v1/internal/products/sync", body)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("AcceptsDescriptionAndTags", func(t *testing.T) {
		var savedProducts []*entity.Product
		repo := &fakeProductRepo{
			saveBatchFunc: func(_ context.Context, products []*entity.Product) error {
				savedProducts = products
				return nil
			},
		}
		uc := usecase.NewCatalogUsecases(repo)
		h := handler.NewProductHandler(uc)
		r, v1 := setupTestRouter()
		v1.POST("/internal/products/sync", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.SyncFromIA(c)
		})

		body := gin.H{
			"products": []gin.H{{
				"user_id":     "user-123",
				"name":        "PLA Blanco",
				"price":       5500,
				"description": "Ideal para principiantes.",
				"tags":        "pla,blanco,1kg",
			}},
		}
		w := performRequest(r, "POST", "/api/v1/internal/products/sync", body)

		assert.Equal(t, http.StatusCreated, w.Code)
		require.Len(t, savedProducts, 1)
		assert.Equal(t, "Ideal para principiantes.", savedProducts[0].Description)
		assert.Equal(t, "pla,blanco,1kg", savedProducts[0].Tags)
	})

	t.Run("EmptyProductsListReturns200", func(t *testing.T) {
		h := handler.NewProductHandler(usecase.NewCatalogUsecases(&fakeProductRepo{}))
		r, v1 := setupTestRouter()
		v1.POST("/internal/products/sync", func(c *gin.Context) {
			c.Set("tenantUserID", "user-123")
			h.SyncFromIA(c)
		})

		w := performRequest(r, "POST", "/api/v1/internal/products/sync", gin.H{"products": []gin.H{}})
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
