package port

import (
	"context"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type PaymentRepository interface {
	Save(ctx context.Context, payment *entity.Payment) error
	ListByUserID(ctx context.Context, userID string) ([]*entity.Payment, error)
	SaveBatch(ctx context.Context, payments []*entity.Payment) error
}
