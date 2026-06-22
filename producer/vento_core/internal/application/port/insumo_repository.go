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
}
