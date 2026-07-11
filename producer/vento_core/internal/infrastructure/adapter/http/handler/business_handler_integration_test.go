//go:build integration

package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/postgres"
)

// Run with: go test -tags integration ./internal/infrastructure/adapter/http/handler/...
// Required env: TEST_DB_URL=postgres://vento:vento_secret@localhost:5433/vento_db?sslmode=disable

func setupBusinessHandlerTestEnv(t *testing.T) *gin.Engine {
	t.Helper()
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set")
	}
	db, err := postgres.NewConnection(dsn)
	require.NoError(t, err)

	userID := "user-integ-" + uuid.New().String()[:8]
	_, err = db.Exec(
		`INSERT INTO users (id, email, password, full_name) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (id) DO NOTHING`,
		userID, userID+"@test.com", "hashed_password", "Test User",
	)
	require.NoError(t, err, "failed to create test user")

	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID)
		db.Close()
	})

	repo := postgres.NewBusinessRepo(db)
	usecases := usecase.NewBusinessProfileUsecases(repo)
	h := handler.NewBusinessHandler(usecases)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	r.PATCH("/business-profile/tone", h.UpdateTone)
	return r
}

func TestIntegration_UpdateTone_ValidValueReturnsOK(t *testing.T) {
	r := setupBusinessHandlerTestEnv(t)

	body, _ := json.Marshal(map[string]string{"tone": "rioplatense"})
	req := httptest.NewRequest(http.MethodPatch, "/business-profile/tone", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "rioplatense", resp["tone"])
}

func TestIntegration_UpdateTone_InvalidValueReturns400(t *testing.T) {
	r := setupBusinessHandlerTestEnv(t)

	body, _ := json.Marshal(map[string]string{"tone": "grosero"})
	req := httptest.NewRequest(http.MethodPatch, "/business-profile/tone", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
