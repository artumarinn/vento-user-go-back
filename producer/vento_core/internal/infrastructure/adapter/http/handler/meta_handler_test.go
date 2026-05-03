package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
)

type fakeMetaRepo struct {
	saveFunc func(ctx context.Context, config *entity.MetaConfig) error
	getByPlatformIDFunc func(ctx context.Context, platformID string) (*entity.MetaConfig, error)
	getByUserIDFunc func(ctx context.Context, userID string) (*entity.MetaConfig, error)
	getAllByUserIDFunc func(ctx context.Context, userID string) ([]*entity.MetaConfig, error)
	getByUserIDAndChannelFunc func(ctx context.Context, userID, channel string) (*entity.MetaConfig, error)
}

func (f *fakeMetaRepo) Save(ctx context.Context, c *entity.MetaConfig) error { return f.saveFunc(ctx, c) }
func (f *fakeMetaRepo) GetByUserID(ctx context.Context, u string) (*entity.MetaConfig, error) { return f.getByUserIDFunc(ctx, u) }
func (f *fakeMetaRepo) GetAllByUserID(ctx context.Context, u string) ([]*entity.MetaConfig, error) { return f.getAllByUserIDFunc(ctx, u) }
func (f *fakeMetaRepo) GetByUserIDAndChannel(ctx context.Context, u, c string) (*entity.MetaConfig, error) { return f.getByUserIDAndChannelFunc(ctx, u, c) }
func (f *fakeMetaRepo) GetByPlatformID(ctx context.Context, p string) (*entity.MetaConfig, error) { return f.getByPlatformIDFunc(ctx, p) }

func TestMetaHandler_Save(t *testing.T) {
	t.Run("Happy Path - Instagram", func(t *testing.T) {
		repo := &fakeMetaRepo{
			saveFunc: func(ctx context.Context, config *entity.MetaConfig) error {
				return nil
			},
		}
		h := handler.NewMetaHandler(repo)
		r, v1 := setupTestRouter()
		v1.POST("/meta-configs", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.Save(c)
		})

		body := gin.H{
			"platform_id":            "insta-1",
			"channel":                "instagram",
			"permanent_access_token": "token-xyz",
		}
		w := performRequest(r, "POST", "/api/v1/meta-configs", body)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Validation Error - Missing PlatformID", func(t *testing.T) {
		h := handler.NewMetaHandler(&fakeMetaRepo{})
		r, v1 := setupTestRouter()
		v1.POST("/meta-configs", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.Save(c)
		})

		body := gin.H{
			"channel":                "instagram",
			"permanent_access_token": "token-xyz",
		}
		w := performRequest(r, "POST", "/api/v1/meta-configs", body)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation Error - Invalid Channel", func(t *testing.T) {
		h := handler.NewMetaHandler(&fakeMetaRepo{})
		r, v1 := setupTestRouter()
		v1.POST("/meta-configs", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.Save(c)
		})

		body := gin.H{
			"platform_id":            "insta-1",
			"channel":                "tiktok",
			"permanent_access_token": "token-xyz",
		}
		w := performRequest(r, "POST", "/api/v1/meta-configs", body)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DB Error", func(t *testing.T) {
		repo := &fakeMetaRepo{
			saveFunc: func(ctx context.Context, config *entity.MetaConfig) error {
				return errors.New("db failure")
			},
		}
		h := handler.NewMetaHandler(repo)
		r, v1 := setupTestRouter()
		v1.POST("/meta-configs", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.Save(c)
		})

		body := gin.H{
			"platform_id":            "insta-1",
			"channel":                "instagram",
			"permanent_access_token": "token-xyz",
		}
		w := performRequest(r, "POST", "/api/v1/meta-configs", body)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
