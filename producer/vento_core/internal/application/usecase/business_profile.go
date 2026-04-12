package usecase

import (
	"context"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/postgres"
)

type BusinessProfileUsecases struct {
	repo *postgres.BusinessRepo
}

func NewBusinessProfileUsecases(repo *postgres.BusinessRepo) *BusinessProfileUsecases {
	return &BusinessProfileUsecases{repo: repo}
}

func (uc *BusinessProfileUsecases) SaveProfile(ctx context.Context, profile *entity.BusinessProfile) error {
	return uc.repo.Save(ctx, profile)
}

func (uc *BusinessProfileUsecases) GetProfile(ctx context.Context, userID string) (*entity.BusinessProfile, error) {
	return uc.repo.GetByUserID(ctx, userID)
}

func (uc *BusinessProfileUsecases) GetStats(ctx context.Context, userID string) (*entity.BusinessStats, error) {
	return uc.repo.GetStats(ctx, userID)
}
