package handler_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
)

// TestLocationHandler_Create_RejectsMissingName guards WARNING-6: an empty
// body must not silently create a nameless location.
func TestLocationHandler_Create_RejectsMissingName(t *testing.T) {
	uc := usecase.NewLocationUsecases(&fakeLocationRepo{})
	h := handler.NewLocationHandler(uc)
	r, v1 := setupTestRouter()
	v1.POST("/locations", func(c *gin.Context) {
		c.Set("userID", "user-123")
		h.Create(c)
	})

	w := performRequest(r, "POST", "/api/v1/locations", gin.H{})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLocationHandler_Create_HappyPath(t *testing.T) {
	uc := usecase.NewLocationUsecases(&fakeLocationRepo{})
	h := handler.NewLocationHandler(uc)
	r, v1 := setupTestRouter()
	v1.POST("/locations", func(c *gin.Context) {
		c.Set("userID", "user-123")
		h.Create(c)
	})

	w := performRequest(r, "POST", "/api/v1/locations", gin.H{"name": "Sucursal Norte"})

	assert.Equal(t, http.StatusCreated, w.Code)
}
