package port

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// ServiceRepository defines the expected behavior for service persistence.
type ServiceRepository interface {
	Save(ctx context.Context, service *entity.Service) error
	Update(ctx context.Context, service *entity.Service) error
	Delete(ctx context.Context, id string, userID string) error
	GetByID(ctx context.Context, id string, userID string) (*entity.Service, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.Service, error)
	SaveBatch(ctx context.Context, services []*entity.Service) error
}
