package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
)

type fakeOrderRepo struct {
	saveFunc func(ctx context.Context, order *entity.Order) error
	getByIDFunc func(ctx context.Context, id string) (*entity.Order, error)
	listByUserIDFunc func(ctx context.Context, userID string) ([]*entity.Order, error)
	updateStatusFunc func(ctx context.Context, id string, status entity.OrderStatus) error
	saveBatchFunc func(ctx context.Context, orders []*entity.Order) error
}

func (f *fakeOrderRepo) Save(ctx context.Context, o *entity.Order) error { return f.saveFunc(ctx, o) }
func (f *fakeOrderRepo) GetByID(ctx context.Context, id string) (*entity.Order, error) { return f.getByIDFunc(ctx, id) }
func (f *fakeOrderRepo) ListByUserID(ctx context.Context, u string) ([]*entity.Order, error) { return f.listByUserIDFunc(ctx, u) }
func (f *fakeOrderRepo) UpdateStatus(ctx context.Context, id string, s entity.OrderStatus) error { return f.updateStatusFunc(ctx, id, s) }
func (f *fakeOrderRepo) SaveBatch(ctx context.Context, o []*entity.Order) error { return f.saveBatchFunc(ctx, o) }

func TestOrderHandler_List(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		repo := &fakeOrderRepo{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]*entity.Order, error) {
				return []*entity.Order{}, nil
			},
		}
		uc := usecase.NewOrderUsecases(repo)
		h := handler.NewOrderHandler(uc)
		r, v1 := setupTestRouter()
		v1.GET("/orders", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.List(c)
		})

		w := performRequest(r, "GET", "/api/v1/orders", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
