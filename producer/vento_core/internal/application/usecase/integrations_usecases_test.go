package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type fakeMetaRepoForIntegrations struct {
	getAllByUserIDFunc func(ctx context.Context, userID string) ([]*entity.MetaConfig, error)
}

func (f *fakeMetaRepoForIntegrations) Save(ctx context.Context, c *entity.MetaConfig) error {
	return nil
}
func (f *fakeMetaRepoForIntegrations) GetByUserID(ctx context.Context, u string) (*entity.MetaConfig, error) {
	return nil, nil
}
func (f *fakeMetaRepoForIntegrations) GetAllByUserID(ctx context.Context, u string) ([]*entity.MetaConfig, error) {
	return f.getAllByUserIDFunc(ctx, u)
}
func (f *fakeMetaRepoForIntegrations) GetByUserIDAndChannel(ctx context.Context, u, c string) (*entity.MetaConfig, error) {
	return nil, nil
}
func (f *fakeMetaRepoForIntegrations) GetByPlatformID(ctx context.Context, p string) (*entity.MetaConfig, error) {
	return nil, nil
}

func TestIntegrationsUsecases_GetStatus(t *testing.T) {
	t.Run("no channels connected", func(t *testing.T) {
		repo := &fakeMetaRepoForIntegrations{
			getAllByUserIDFunc: func(ctx context.Context, u string) ([]*entity.MetaConfig, error) {
				return []*entity.MetaConfig{}, nil
			},
		}
		uc := usecase.NewIntegrationsUsecases(repo)

		status, err := uc.GetStatus(context.Background(), "user-123")

		assert.NoError(t, err)
		assert.False(t, status.WhatsApp)
		assert.False(t, status.Instagram)
		assert.False(t, status.Facebook)
		assert.False(t, status.MercadoPago)
		assert.Equal(t, 0, status.Connected)
		assert.Equal(t, 4, status.Total)
	})

	t.Run("whatsapp and instagram connected", func(t *testing.T) {
		repo := &fakeMetaRepoForIntegrations{
			getAllByUserIDFunc: func(ctx context.Context, u string) ([]*entity.MetaConfig, error) {
				return []*entity.MetaConfig{
					{Channel: "whatsapp"},
					{Channel: "instagram"},
				}, nil
			},
		}
		uc := usecase.NewIntegrationsUsecases(repo)

		status, err := uc.GetStatus(context.Background(), "user-123")

		assert.NoError(t, err)
		assert.True(t, status.WhatsApp)
		assert.True(t, status.Instagram)
		assert.False(t, status.Facebook)
		assert.False(t, status.MercadoPago)
		assert.Equal(t, 2, status.Connected)
		assert.Equal(t, 4, status.Total)
	})

	t.Run("facebook channel is stored as messenger", func(t *testing.T) {
		repo := &fakeMetaRepoForIntegrations{
			getAllByUserIDFunc: func(ctx context.Context, u string) ([]*entity.MetaConfig, error) {
				return []*entity.MetaConfig{
					{Channel: "messenger"},
				}, nil
			},
		}
		uc := usecase.NewIntegrationsUsecases(repo)

		status, err := uc.GetStatus(context.Background(), "user-123")

		assert.NoError(t, err)
		assert.True(t, status.Facebook)
		assert.Equal(t, 1, status.Connected)
	})

	t.Run("repo error is propagated", func(t *testing.T) {
		repo := &fakeMetaRepoForIntegrations{
			getAllByUserIDFunc: func(ctx context.Context, u string) ([]*entity.MetaConfig, error) {
				return nil, errors.New("db failure")
			},
		}
		uc := usecase.NewIntegrationsUsecases(repo)

		_, err := uc.GetStatus(context.Background(), "user-123")

		assert.Error(t, err)
	})
}
