package port

import (
	"context"
	"time"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// UserRepository defines the persistence port for User entities.
// Infrastructure adapters (Postgres, Mongo, etc.) implement this interface.
type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*entity.User, error)
	UpdatePassword(ctx context.Context, email, hashedPassword string) error

	// Password Reset Tokens
	SaveResetToken(ctx context.Context, email, token string, expiresAt time.Time) error
	GetEmailByResetToken(ctx context.Context, token string) (string, error)
	DeleteResetToken(ctx context.Context, token string) error
}

