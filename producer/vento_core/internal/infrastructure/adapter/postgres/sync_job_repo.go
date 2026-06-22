package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type PostgresSyncJobRepository struct {
	db *sqlx.DB
}

func NewPostgresSyncJobRepository(db *sqlx.DB) *PostgresSyncJobRepository {
	return &PostgresSyncJobRepository{db: db}
}

func (r *PostgresSyncJobRepository) Create(ctx context.Context, job *entity.SyncJob) error {
	query := `
		INSERT INTO sync_jobs (id, user_id, type, status, progress, error_msg, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		job.ID, job.UserID, job.Type, job.Status, job.Progress, job.ErrorMsg, job.CreatedAt, job.UpdatedAt,
	)
	return err
}

func (r *PostgresSyncJobRepository) GetByID(ctx context.Context, id string) (*entity.SyncJob, error) {
	var job entity.SyncJob
	query := `SELECT id, user_id, type, status, progress, error_msg, created_at, updated_at FROM sync_jobs WHERE id = $1`
	err := r.db.GetContext(ctx, &job, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("sync job not found")
		}
		return nil, err
	}
	return &job, nil
}

func (r *PostgresSyncJobRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.SyncJob, error) {
	var jobs []*entity.SyncJob
	query := `SELECT id, user_id, type, status, progress, error_msg, created_at, updated_at FROM sync_jobs WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &jobs, query, userID)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *PostgresSyncJobRepository) UpdateProgress(ctx context.Context, id string, progress int, status string, errorMsg *string) error {
	query := `
		UPDATE sync_jobs
		SET progress = $1, status = $2, error_msg = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, progress, status, errorMsg, id)
	return err
}

func (r *PostgresSyncJobRepository) GetNextQueuedJob(ctx context.Context) (*entity.SyncJob, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var job entity.SyncJob
	selectQuery := `
		SELECT id, user_id, type, status, progress, error_msg, created_at, updated_at
		FROM sync_jobs
		WHERE status = 'QUEUED'
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`
	err = tx.GetContext(ctx, &job, selectQuery)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No jobs queued
		}
		return nil, err
	}

	updateQuery := `
		UPDATE sync_jobs
		SET status = 'PROCESSING', updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, updateQuery, job.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Update returning model status to PROCESSING before returning
	job.Status = entity.SyncJobStatusProcessing
	return &job, nil
}

