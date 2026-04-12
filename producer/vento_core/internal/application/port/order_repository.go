package port

import (
	"context"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type OrderRepository interface {
	Save(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id string) (*entity.Order, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.Order, error)
	UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error
	SaveBatch(ctx context.Context, orders []*entity.Order) error
}
