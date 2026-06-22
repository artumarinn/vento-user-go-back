package port

import (
	"context"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type SyncJobRepository interface {
	Create(ctx context.Context, job *entity.SyncJob) error
	GetByID(ctx context.Context, id string) (*entity.SyncJob, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.SyncJob, error)
	UpdateProgress(ctx context.Context, id string, progress int, status string, errorMsg *string) error
	GetNextQueuedJob(ctx context.Context) (*entity.SyncJob, error)
}
