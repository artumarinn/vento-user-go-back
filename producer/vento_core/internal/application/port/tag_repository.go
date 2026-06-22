package port

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// TagRepository defines the expected behavior for tag persistence.
type TagRepository interface {
	Save(ctx context.Context, tag *entity.Tag) error
	Delete(ctx context.Context, id string, userID string) error
	ListByUserID(ctx context.Context, userID string) ([]*entity.Tag, error)
	ListByInsumoID(ctx context.Context, insumoID string) ([]*entity.Tag, error)
	SetInsumoTags(ctx context.Context, insumoID string, tagIDs []string) error
	ListByProductID(ctx context.Context, productID string) ([]*entity.Tag, error)
	SetProductTags(ctx context.Context, productID string, tagIDs []string) error
	ListByServiceID(ctx context.Context, serviceID string) ([]*entity.Tag, error)
	SetServiceTags(ctx context.Context, serviceID string, tagIDs []string) error
}
