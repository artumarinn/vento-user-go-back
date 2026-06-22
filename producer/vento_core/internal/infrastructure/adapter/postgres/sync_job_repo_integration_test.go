package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/postgres"
)

type syncJobTestEnv struct {
	repo   *postgres.PostgresSyncJobRepository
	db     *sqlx.DB
	userID string
}

func setupSyncJobTestEnv(t *testing.T) *syncJobTestEnv {
	t.Helper()
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set")
	}

	db, err := postgres.NewConnection(dsn)
	require.NoError(t, err)

	userID := "user-sync-" + uuid.New().String()[:8]
	_, err = db.Exec(
		`INSERT INTO users (id, email, password, full_name) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (id) DO NOTHING`,
		userID,
		userID+"@test.com",
		"hashed_password",
		"Test User",
	)
	require.NoError(t, err, "failed to create test user")

	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM sync_jobs WHERE user_id = $1`, userID)
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID)
		db.Close()
	})

	return &syncJobTestEnv{
		repo:   postgres.NewPostgresSyncJobRepository(db),
		db:     db,
		userID: userID,
	}
}

func TestIntegration_SyncJobRepo_GetNextQueuedJob(t *testing.T) {
	env := setupSyncJobTestEnv(t)
	ctx := context.Background()

	// 1. Initially there should be no queued jobs
	job, err := env.repo.GetNextQueuedJob(ctx)
	require.NoError(t, err)
	assert.Nil(t, job)

	// 2. Insert two queued jobs
	job1 := &entity.SyncJob{
		ID:        uuid.New().String(),
		UserID:    env.userID,
		Type:      "PRODUCTS",
		Status:    entity.SyncJobStatusQueued,
		Progress:  0,
		CreatedAt: time.Now().Add(-10 * time.Minute),
		UpdatedAt: time.Now().Add(-10 * time.Minute),
	}
	job2 := &entity.SyncJob{
		ID:        uuid.New().String(),
		UserID:    env.userID,
		Type:      "ORDERS",
		Status:    entity.SyncJobStatusQueued,
		Progress:  0,
		CreatedAt: time.Now().Add(-5 * time.Minute),
		UpdatedAt: time.Now().Add(-5 * time.Minute),
	}

	require.NoError(t, env.repo.Create(ctx, job1))
	require.NoError(t, env.repo.Create(ctx, job2))

	// 3. Get next queued job - should return job1 (older) and transition to PROCESSING
	fetchedJob, err := env.repo.GetNextQueuedJob(ctx)
	require.NoError(t, err)
	require.NotNil(t, fetchedJob)
	assert.Equal(t, job1.ID, fetchedJob.ID)
	assert.Equal(t, entity.SyncJobStatusProcessing, fetchedJob.Status)

	// Verify status in DB is PROCESSING
	dbJob1, err := env.repo.GetByID(ctx, job1.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.SyncJobStatusProcessing, dbJob1.Status)

	// 4. Get next queued job again - should return job2 (newer)
	fetchedJob2, err := env.repo.GetNextQueuedJob(ctx)
	require.NoError(t, err)
	require.NotNil(t, fetchedJob2)
	assert.Equal(t, job2.ID, fetchedJob2.ID)
	assert.Equal(t, entity.SyncJobStatusProcessing, fetchedJob2.Status)

	// 5. Get next queued job again - none should be left
	fetchedJob3, err := env.repo.GetNextQueuedJob(ctx)
	require.NoError(t, err)
	assert.Nil(t, fetchedJob3)
}
