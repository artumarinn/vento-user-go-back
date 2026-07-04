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
	AdjustStock(ctx context.Context, productID string, userID string, delta float64) error
	InsertStockMovement(ctx context.Context, movement *entity.StockMovement) error
	// AdjustLocationStock upserts location_stock, applying delta to a
	// product's quantity at a specific location. A no-op when locationID is
	// empty (order/context without a resolved location).
	AdjustLocationStock(ctx context.Context, locationID string, productID string, delta float64) error
	// GetLocationStock returns the current quantity of a product at a
	// specific location. Returns 0 when no row exists yet.
	GetLocationStock(ctx context.Context, locationID string, productID string) (float64, error)
}
