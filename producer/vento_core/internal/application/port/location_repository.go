package port

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// LocationRepository defines the expected behavior for location persistence.
type LocationRepository interface {
	Save(ctx context.Context, location *entity.Location) error
	Update(ctx context.Context, location *entity.Location) error
	Delete(ctx context.Context, id string, userID string) error
	GetByID(ctx context.Context, id string, userID string) (*entity.Location, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.Location, error)
	GetDefault(ctx context.Context, userID string) (*entity.Location, error)
	CountByUserID(ctx context.Context, userID string) (int, error)
	// UnsetDefault clears is_default on every location of a user except the
	// one whose ID is given. Used when promoting a new default location.
	UnsetDefault(ctx context.Context, userID string, exceptID string) error
}
