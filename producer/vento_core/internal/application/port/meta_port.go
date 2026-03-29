package port

import (
	"context"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type MetaRepository interface {
	Save(ctx context.Context, config *entity.MetaConfig) error
	GetByUserID(ctx context.Context, userID string) (*entity.MetaConfig, error)
	GetByPhoneNumberID(ctx context.Context, phoneNumberID string) (*entity.MetaConfig, error)
}
