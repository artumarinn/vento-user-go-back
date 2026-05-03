package port

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type MetaRepository interface {
	Save(ctx context.Context, config *entity.MetaConfig) error
	GetByUserID(ctx context.Context, userID string) (*entity.MetaConfig, error)
	GetAllByUserID(ctx context.Context, userID string) ([]*entity.MetaConfig, error)
	GetByUserIDAndChannel(ctx context.Context, userID, channel string) (*entity.MetaConfig, error)
	GetByPlatformID(ctx context.Context, platformID string) (*entity.MetaConfig, error)
}
