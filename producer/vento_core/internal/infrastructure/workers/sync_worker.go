package workers

import (
	"context"
	"time"

	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type SyncWorker struct {
	repo          port.SyncJobRepository
	pollInterval  time.Duration
}

func NewSyncWorker(repo port.SyncJobRepository) *SyncWorker {
	return &SyncWorker{
		repo:         repo,
		pollInterval: 3 * time.Second,
	}
}

func (w *SyncWorker) Start(ctx context.Context) {
	logger.L().Info("Background SyncWorker started")
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.L().Info("Stopping Background SyncWorker gracefully...")
			return
		case <-ticker.C:
			w.pollAndProcess(ctx)
		}
	}
}

func (w *SyncWorker) pollAndProcess(ctx context.Context) {
	job, err := w.repo.GetNextQueuedJob(ctx)
	if err != nil {
		logger.L().Error("SyncWorker failed to fetch next queued job", "error", err)
		return
	}
	if job == nil {
		return
	}

	logger.L().Info("SyncWorker starting job processing", "job_id", job.ID, "user_id", job.UserID, "type", job.Type)

	// Mock process it: sleep 2 seconds
	select {
	case <-ctx.Done():
		logger.L().Warn("SyncWorker context cancelled during job processing, rolling back progress update in DB", "job_id", job.ID)
		errMsg := "Worker stopped during execution"
		_ = w.repo.UpdateProgress(context.Background(), job.ID, job.Progress, entity.SyncJobStatusFailed, &errMsg)
		return
	case <-time.After(2 * time.Second):
	}

	// Mock processing logic: we succeed
	logger.L().Info("SyncWorker successfully processed job", "job_id", job.ID)
	err = w.repo.UpdateProgress(ctx, job.ID, 100, entity.SyncJobStatusSuccess, nil)
	if err != nil {
		logger.L().Error("SyncWorker failed to update job progress to success in DB", "job_id", job.ID, "error", err)
	}
}
