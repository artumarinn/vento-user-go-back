package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
)

func TestIntegrationsHandler_GetStatus(t *testing.T) {
	t.Run("Happy Path - some channels connected", func(t *testing.T) {
		repo := &fakeMetaRepo{
			getAllByUserIDFunc: func(ctx context.Context, u string) ([]*entity.MetaConfig, error) {
				return []*entity.MetaConfig{
					{Channel: "whatsapp"},
				}, nil
			},
		}
		uc := usecase.NewIntegrationsUsecases(repo)
		h := handler.NewIntegrationsHandler(uc)
		r, v1 := setupTestRouter()
		v1.GET("/integrations/status", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.GetStatus(c)
		})

		w := performRequest(r, "GET", "/api/v1/integrations/status", nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var body map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &body)
		assert.NoError(t, err)
		assert.Equal(t, true, body["whatsapp"])
		assert.Equal(t, false, body["instagram"])
		assert.Equal(t, false, body["facebook"])
		assert.Equal(t, false, body["mercadopago"])
		assert.Equal(t, float64(1), body["connected"])
		assert.Equal(t, float64(4), body["total"])
	})

	t.Run("Repo Error", func(t *testing.T) {
		repo := &fakeMetaRepo{
			getAllByUserIDFunc: func(ctx context.Context, u string) ([]*entity.MetaConfig, error) {
				return nil, errors.New("db failure")
			},
		}
		uc := usecase.NewIntegrationsUsecases(repo)
		h := handler.NewIntegrationsHandler(uc)
		r, v1 := setupTestRouter()
		v1.GET("/integrations/status", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.GetStatus(c)
		})

		w := performRequest(r, "GET", "/api/v1/integrations/status", nil)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
