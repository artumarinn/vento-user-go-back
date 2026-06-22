package handler_test

import (
	"context"
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

type seededServiceRepo struct {
	services map[string]*entity.Service
}

func (f *seededServiceRepo) Save(ctx context.Context, service *entity.Service) error   { return nil }
func (f *seededServiceRepo) Update(ctx context.Context, service *entity.Service) error { return nil }
func (f *seededServiceRepo) Delete(ctx context.Context, id string, userID string) error {
	return nil
}
func (f *seededServiceRepo) GetByID(ctx context.Context, id string, userID string) (*entity.Service, error) {
	return f.services[id], nil
}
func (f *seededServiceRepo) ListByUserID(ctx context.Context, userID string) ([]*entity.Service, error) {
	return nil, nil
}
func (f *seededServiceRepo) SaveBatch(ctx context.Context, services []*entity.Service) error {
	return nil
}

func ptr64(f float64) *float64 { return &f }
func ptrS(s string) *string    { return &s }

func seededPrintingService() *entity.Service {
	s := entity.Service(catalog.Service{
		ID:      "svc-1",
		Name:    "Impresión 3D",
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
	})
	return &s
}

func TestServiceHandler_PricePreview_Success(t *testing.T) {
	repo := &seededServiceRepo{services: map[string]*entity.Service{
		"svc-1": seededPrintingService(),
	}}
	uc := usecase.NewServiceUsecases(repo, nil)
	h := handler.NewServiceHandler(uc)

	r, v1 := setupTestRouter()
	v1.POST("/services/:id/price-preview", func(c *gin.Context) {
		c.Set("userID", "user-1")
		h.PricePreview(c)
	})

	body := dto.PreviewPriceRequest{
		Variables: []entity.OrderItemVariable{
			{Name: "material", Type: "select", OptionValue: ptrS("PLA")},
			{Name: "peso", Type: "number", NumberValue: ptr64(150)},
		},
	}

	w := performRequest(r, "POST", "/api/v1/services/svc-1/price-preview", body)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"unit_price":300`)
}

func TestServiceHandler_PricePreview_EvaluatorErrorReturns422(t *testing.T) {
	repo := &seededServiceRepo{services: map[string]*entity.Service{
		"svc-1": seededPrintingService(),
	}}
	uc := usecase.NewServiceUsecases(repo, nil)
	h := handler.NewServiceHandler(uc)

	r, v1 := setupTestRouter()
	v1.POST("/services/:id/price-preview", func(c *gin.Context) {
		c.Set("userID", "user-1")
		h.PricePreview(c)
	})

	body := dto.PreviewPriceRequest{
		Variables: []entity.OrderItemVariable{
			{Name: "material", Type: "select", OptionValue: ptrS("PVC")},
			{Name: "peso", Type: "number", NumberValue: ptr64(150)},
		},
	}

	w := performRequest(r, "POST", "/api/v1/services/svc-1/price-preview", body)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
