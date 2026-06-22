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

type fakePaymentRepo struct {
	saveFunc func(ctx context.Context, payment *entity.Payment) error
	listByUserIDFunc func(ctx context.Context, userID string) ([]*entity.Payment, error)
	saveBatchFunc func(ctx context.Context, payments []*entity.Payment) error
}

func (f *fakePaymentRepo) Save(ctx context.Context, p *entity.Payment) error { return f.saveFunc(ctx, p) }
func (f *fakePaymentRepo) ListByUserID(ctx context.Context, u string) ([]*entity.Payment, error) { return f.listByUserIDFunc(ctx, u) }
func (f *fakePaymentRepo) SaveBatch(ctx context.Context, p []*entity.Payment) error { return f.saveBatchFunc(ctx, p) }

func TestPaymentHandler_List(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		repo := &fakePaymentRepo{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]*entity.Payment, error) {
				return []*entity.Payment{}, nil
			},
		}
		uc := usecase.NewPaymentUsecases(repo, nil)
		h := handler.NewPaymentHandler(uc)
		r, v1 := setupTestRouter()
		v1.GET("/payments", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.List(c)
		})

		w := performRequest(r, "GET", "/api/v1/payments", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
