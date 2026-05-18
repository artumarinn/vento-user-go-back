//go:build integration

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

// Run with: go test -tags integration ./internal/infrastructure/adapter/postgres/...
// Required env: TEST_DB_URL=postgres://vento:vento_secret@localhost:5433/vento_db?sslmode=disable

type productTestEnv struct {
	repo   *postgres.PostgresProductRepository
	db     *sqlx.DB
	userID string
}

func setupProductTestEnv(t *testing.T) *productTestEnv {
	t.Helper()
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set")
	}

	// NewConnection applies all migrations including description/tags columns
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
		// Products cascade-delete with user, so this cleans everything
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID)
		db.Close()
	})

	return &productTestEnv{
		repo:   postgres.NewPostgresProductRepository(db),
		db:     db,
		userID: userID,
	}
}

func newTestProduct(userID string) *entity.Product {
	p := entity.NewProduct(userID, "Test PLA "+uuid.New().String()[:8], "SKU-TEST", "Filamento PLA", 5500)
	p.Stock = 8.5
	p.StockUnit = "kg"
	return p
}

func TestIntegration_ProductRepo_SaveAndRetrieveDescriptionAndTags(t *testing.T) {
	env := setupProductTestEnv(t)
	ctx := context.Background()

	p := newTestProduct(env.userID)
	p.Description = "Filamento ideal para principiantes."
	p.Tags = "pla,filamento,blanco"

	require.NoError(t, env.repo.Save(ctx, p))

	products, err := env.repo.ListByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.Len(t, products, 1)
	assert.Equal(t, "Filamento ideal para principiantes.", products[0].Description)
	assert.Equal(t, "pla,filamento,blanco", products[0].Tags)
}

func TestIntegration_ProductRepo_OnConflict_PreservesDescriptionWhenEmpty(t *testing.T) {
	env := setupProductTestEnv(t)
	ctx := context.Background()

	p := newTestProduct(env.userID)
	p.Description = "Descripción original no debe perderse."
	p.Tags = "tag1,tag2"
	require.NoError(t, env.repo.Save(ctx, p))

	// Stock update with empty description — should NOT erase existing description
	p2 := *p
	p2.Stock = 3.0
	p2.Description = ""
	p2.Tags = ""
	p2.UpdatedAt = time.Now()
	require.NoError(t, env.repo.Save(ctx, &p2))

	products, err := env.repo.ListByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.Len(t, products, 1)
	assert.Equal(t, "Descripción original no debe perderse.", products[0].Description,
		"description must be preserved when empty string is saved on conflict")
	assert.Equal(t, "tag1,tag2", products[0].Tags)
	assert.Equal(t, float64(3), products[0].Stock, "stock must be updated")
}

func TestIntegration_ProductRepo_OnConflict_UpdatesDescriptionWhenProvided(t *testing.T) {
	env := setupProductTestEnv(t)
	ctx := context.Background()

	p := newTestProduct(env.userID)
	p.Description = "Descripción inicial."
	require.NoError(t, env.repo.Save(ctx, p))

	p.Description = "Descripción actualizada."
	p.UpdatedAt = time.Now()
	require.NoError(t, env.repo.Save(ctx, p))

	products, err := env.repo.ListByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.Len(t, products, 1)
	assert.Equal(t, "Descripción actualizada.", products[0].Description)
}

func TestIntegration_ProductRepo_SaveBatch_PersistsAllFields(t *testing.T) {
	env := setupProductTestEnv(t)
	ctx := context.Background()

	products := []*entity.Product{
		{
			ID: uuid.New().String(), UserID: env.userID, Name: "PLA Blanco " + uuid.New().String()[:4],
			Category: "Filamento PLA", Stock: 8.5, StockUnit: "kg", Price: 5500,
			Description: "Desc A", Tags: "pla,blanco", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		{
			ID: uuid.New().String(), UserID: env.userID, Name: "ABS Negro " + uuid.New().String()[:4],
			Category: "Filamento ABS", Stock: 4.0, StockUnit: "kg", Price: 6000,
			Description: "Desc B", Tags: "abs,negro", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}
	require.NoError(t, env.repo.SaveBatch(ctx, products))

	result, err := env.repo.ListByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.Len(t, result, 2)

	descMap := map[string]string{}
	for _, p := range result {
		descMap[p.Name] = p.Description
	}
	assert.Equal(t, "Desc A", descMap[products[0].Name])
	assert.Equal(t, "Desc B", descMap[products[1].Name])
}
