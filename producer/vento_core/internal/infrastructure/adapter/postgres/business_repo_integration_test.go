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

// TestIntegration_BusinessRepo_Save_FreshProfileGetsDefaultAgentMode confirms
// Save() no longer accepts DefaultAgentMode on the input at all (see
// TestIntegration_BusinessRepo_Save_DoesNotTouchAgentMode) — a brand-new
// profile gets the column's own DB default ('autonomous'), and the mode is
// only ever changed via UpdateAgentMode.
func TestIntegration_BusinessRepo_Save_FreshProfileGetsDefaultAgentMode(t *testing.T) {
	env := setupBusinessTestEnv(t)
	ctx := context.Background()

	profile := &entity.BusinessProfile{
		UserID:           env.userID,
		BusinessName:     "Impresiones 3D",
		DefaultAgentMode: "assisted", // ignored by Save on purpose
	}

	require.NoError(t, env.repo.Save(ctx, profile))

	loaded, err := env.repo.GetByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.Equal(t, "autonomous", loaded.DefaultAgentMode)
}

// TestIntegration_BusinessRepo_UpdateAgentMode_PreservesOtherFields reproduces
// the data-loss bug: the "Cómo responde" autonomous toggle sent a partial
// payload to the full Save upsert, which blanked every unspecified column —
// wiping the whole business profile every time the mode was flipped.
// UpdateAgentMode must touch ONLY default_agent_mode.
func TestIntegration_BusinessRepo_UpdateAgentMode_PreservesOtherFields(t *testing.T) {
	env := setupBusinessTestEnv(t)
	ctx := context.Background()

	full := &entity.BusinessProfile{
		UserID:           env.userID,
		BusinessName:     "Impre3D",
		Description:      "Estudio de impresión 3D en FDM y resina.",
		Industry:         "Impresión 3D",
		Tone:             "rioplatense", // ignored by Save on purpose, set via UpdateTone below
		Currency:         "ARS",
		DefaultAgentMode: "autonomous", // ignored by Save on purpose
	}
	require.NoError(t, env.repo.Save(ctx, full))
	require.NoError(t, env.repo.UpdateTone(ctx, env.userID, "rioplatense"))

	// Flip ONLY the mode (what the toggle does) — must not wipe anything else.
	require.NoError(t, env.repo.UpdateAgentMode(ctx, env.userID, "assisted"))

	loaded, err := env.repo.GetByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.Equal(t, "assisted", loaded.DefaultAgentMode)
	require.Equal(t, "Impre3D", loaded.BusinessName)
	require.Equal(t, "Estudio de impresión 3D en FDM y resina.", loaded.Description)
	require.Equal(t, "Impresión 3D", loaded.Industry)
	require.Equal(t, "rioplatense", loaded.Tone)
	require.Equal(t, "ARS", loaded.Currency)
}

// TestIntegration_BusinessRepo_Save_DoesNotTouchAgentMode reproduces the
// mirror bug: saving the full profile from "Mi Negocio" (which never sends
// default_agent_mode) blanked the mode to "" because Save's own upsert wrote
// EXCLUDED.default_agent_mode unconditionally. Save must leave that column
// alone — it's owned by UpdateAgentMode now.
func TestIntegration_BusinessRepo_Save_DoesNotTouchAgentMode(t *testing.T) {
	env := setupBusinessTestEnv(t)
	ctx := context.Background()

	require.NoError(t, env.repo.UpdateAgentMode(ctx, env.userID, "assisted"))

	// A full-profile save that never sets DefaultAgentMode (zero value "").
	require.NoError(t, env.repo.Save(ctx, &entity.BusinessProfile{
		UserID:       env.userID,
		BusinessName: "Impre3D",
		Description:  "Estudio de impresión 3D.",
	}))

	loaded, err := env.repo.GetByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.Equal(t, "assisted", loaded.DefaultAgentMode)
	require.Equal(t, "Impre3D", loaded.BusinessName)
}

// TestIntegration_BusinessRepo_UpdateTone_PreservesOtherFields reproduces
// the same protection already proven for UpdateAgentMode: changing the
// tone must never touch business_name/description/industry/currency/
// default_agent_mode.
func TestIntegration_BusinessRepo_UpdateTone_PreservesOtherFields(t *testing.T) {
	env := setupBusinessTestEnv(t)
	ctx := context.Background()

	full := &entity.BusinessProfile{
		UserID:       env.userID,
		BusinessName: "Impre3D",
		Description:  "Estudio de impresión 3D.",
		Industry:     "Impresión 3D",
		Tone:         "profesional",
		Currency:     "ARS",
	}
	require.NoError(t, env.repo.Save(ctx, full))
	require.NoError(t, env.repo.UpdateAgentMode(ctx, env.userID, "assisted"))

	require.NoError(t, env.repo.UpdateTone(ctx, env.userID, "rioplatense"))

	loaded, err := env.repo.GetByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.Equal(t, "rioplatense", loaded.Tone)
	require.Equal(t, "Impre3D", loaded.BusinessName)
	require.Equal(t, "Estudio de impresión 3D.", loaded.Description)
	require.Equal(t, "Impresión 3D", loaded.Industry)
	require.Equal(t, "ARS", loaded.Currency)
	require.Equal(t, "assisted", loaded.DefaultAgentMode)
}

// TestIntegration_BusinessRepo_Save_DoesNotTouchTone mirrors
// TestIntegration_BusinessRepo_Save_DoesNotTouchAgentMode — the mirror bug,
// this time on tone: "Mi Negocio" no longer sends this field, so Save's own
// upsert must not blank it.
func TestIntegration_BusinessRepo_Save_DoesNotTouchTone(t *testing.T) {
	env := setupBusinessTestEnv(t)
	ctx := context.Background()

	require.NoError(t, env.repo.UpdateTone(ctx, env.userID, "rioplatense"))

	require.NoError(t, env.repo.Save(ctx, &entity.BusinessProfile{
		UserID:       env.userID,
		BusinessName: "Impre3D",
	}))

	loaded, err := env.repo.GetByUserID(ctx, env.userID)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.Equal(t, "rioplatense", loaded.Tone)
	require.Equal(t, "Impre3D", loaded.BusinessName)
}
