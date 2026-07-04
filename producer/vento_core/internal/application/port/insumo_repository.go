package port

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// InsumoRepository defines the expected behavior for insumo persistence.
type InsumoRepository interface {
	Save(ctx context.Context, insumo *entity.Insumo) error
	Update(ctx context.Context, insumo *entity.Insumo) error
	Delete(ctx context.Context, id string, userID string) error
	GetByID(ctx context.Context, id string, userID string) (*entity.Insumo, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.Insumo, error)
	SaveBatch(ctx context.Context, insumos []*entity.Insumo) error
	AdjustStock(ctx context.Context, insumoID string, userID string, delta float64) error
	InsertInsumoMovement(ctx context.Context, movement *entity.InsumoMovement) error
	// AdjustLocationInsumoStock upserts location_insumo_stock, applying delta
	// to an insumo's quantity at a specific location. A no-op when
	// locationID is empty.
	AdjustLocationInsumoStock(ctx context.Context, locationID string, insumoID string, delta float64) error
	// GetLocationInsumoStock returns the current quantity of an insumo at a
	// specific location. Returns 0 when no row exists yet.
	GetLocationInsumoStock(ctx context.Context, locationID string, insumoID string) (float64, error)
}
