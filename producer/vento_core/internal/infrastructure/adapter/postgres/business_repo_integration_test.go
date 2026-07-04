//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/postgres"
)

// Run with: go test -tags integration ./internal/infrastructure/adapter/postgres/...
// Required env: TEST_DB_URL=postgres://vento:vento_secret@localhost:5433/vento_db?sslmode=disable

type businessTestEnv struct {
	repo   *postgres.BusinessRepo
	db     *sqlx.DB
	userID string
}

func setupBusinessTestEnv(t *testing.T) *businessTestEnv {
	t.Helper()
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set")
	}

	// NewConnection applies all migrations including default_agent_mode.
	db, err := postgres.NewConnection(dsn)
	require.NoError(t, err)

	// Create a unique test user to satisfy the FK constraint
	userID := "user-integ-" + uuid.New().String()[:8]
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
		// business_profiles cascade-deletes with user, so this cleans everything
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID)
		db.Close()
	})

	return &businessTestEnv{
		repo:   postgres.NewBusinessRepo(db),
		db:     db,
		userID: userID,
	}
}

func TestIntegration_BusinessRepo_DefaultAgentMode_RoundTrips(t *testing.T) {
	env := setupBusinessTestEnv(t)
	ctx := context.Background()

	profile := &entity.BusinessProfile{
		UserID:           env.userID,
		BusinessName:     "Impresiones 3D",
		DefaultAgentMode: "assisted",
	}

	require.NoError(t, env.repo.Save(ctx, profile))

	loaded, err := env.repo.GetByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.Equal(t, "assisted", loaded.DefaultAgentMode)
}
