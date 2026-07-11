package handler_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
)

func TestServiceHandler_PreviewPriceDraft_Success(t *testing.T) {
	uc := usecase.NewServiceUsecases(nil, nil, nil)
	h := handler.NewServiceHandler(uc)

	r, v1 := setupTestRouter()
	v1.POST("/services/preview-price", func(c *gin.Context) {
		h.PreviewPriceDraft(c)
	})

	body := dto.PreviewPriceDraftRequest{
		Formula: "material * peso",
		VariablesSchema: []catalog.VariableDefinition{
			{
				Name:     "material",
				Type:     catalog.VariableTypeSelect,
				Required: true,
				Options: []catalog.VariableOption{
					{Label: "PLA", Value: "PLA", UnitCost: 2},
				},
			},
			{Name: "peso", Type: catalog.VariableTypeNumber, Required: true},
		},
		Variables: []entity.OrderItemVariable{
			{Name: "material", Type: "select", OptionValue: ptrS("PLA")},
			{Name: "peso", Type: "number", NumberValue: ptr64(150)},
		},
	}

	w := performRequest(r, "POST", "/api/v1/services/preview-price", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"unit_price":300`)
}

func TestServiceHandler_PreviewPriceDraft_UnknownVariableReturns422(t *testing.T) {
	uc := usecase.NewServiceUsecases(nil, nil, nil)
	h := handler.NewServiceHandler(uc)

	r, v1 := setupTestRouter()
	v1.POST("/services/preview-price", func(c *gin.Context) {
		h.PreviewPriceDraft(c)
	})

	body := dto.PreviewPriceDraftRequest{
		Formula: "material * peso * tiempo",
		VariablesSchema: []catalog.VariableDefinition{
			{
				Name:     "material",
				Type:     catalog.VariableTypeSelect,
				Required: true,
				Options: []catalog.VariableOption{
					{Label: "PLA", Value: "PLA", UnitCost: 2},
				},
			},
			{Name: "peso", Type: catalog.VariableTypeNumber, Required: true},
		},
		Variables: []entity.OrderItemVariable{
			{Name: "material", Type: "select", OptionValue: ptrS("PLA")},
			{Name: "peso", Type: "number", NumberValue: ptr64(150)},
		},
	}

	w := performRequest(r, "POST", "/api/v1/services/preview-price", body)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
