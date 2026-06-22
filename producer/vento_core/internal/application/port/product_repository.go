package port

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// ProductRepository defines the expected behavior for product persistence.
type ProductRepository interface {
	Save(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id string, userID string) error
	GetByID(ctx context.Context, id string, userID string) (*entity.Product, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.Product, error)
	SaveBatch(ctx context.Context, products []*entity.Product) error
	Search(ctx context.Context, userID string, query string, category string, limit int) ([]*entity.Product, error)
}
