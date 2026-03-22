package port

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// UserRepository defines the persistence port for User entities.
// Infrastructure adapters (Postgres, Mongo, etc.) implement this interface.
type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
}
