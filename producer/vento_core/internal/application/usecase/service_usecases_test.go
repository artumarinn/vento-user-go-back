package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// fakeServiceRepo satisfies port.ServiceRepository for these tests.
type fakeServiceRepo struct {
	services    []*entity.Service
	insumosByID map[string][]entity.ServiceInsumoDetail
}

func (f *fakeServiceRepo) Save(ctx context.Context, s *entity.Service) error   { return nil }
func (f *fakeServiceRepo) Update(ctx context.Context, s *entity.Service) error { return nil }
func (f *fakeServiceRepo) Delete(ctx context.Context, id, userID string) error { return nil }
func (f *fakeServiceRepo) SaveBatch(ctx context.Context, s []*entity.Service) error { return nil }
func (f *fakeServiceRepo) SetServiceInsumos(ctx context.Context, serviceID string, insumos []entity.ServiceInsumo) error {
	return nil
}

func (f *fakeServiceRepo) GetByID(ctx context.Context, id, userID string) (*entity.Service, error) {
	for _, s := range f.services {
		if s.ID == id && s.UserID == userID {
			return s, nil
		}
	}
	return nil, nil
}

func (f *fakeServiceRepo) ListByUserID(ctx context.Context, userID string) ([]*entity.Service, error) {
	var out []*entity.Service
	for _, s := range f.services {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakeServiceRepo) ListServiceInsumos(ctx context.Context, serviceID string) ([]entity.ServiceInsumoDetail, error) {
	return f.insumosByID[serviceID], nil
}

// fakeInsumoRepoForService satisfies port.InsumoRepository for these tests.
type fakeInsumoRepoForService struct {
	insumos map[string]*entity.Insumo
}

func (f *fakeInsumoRepoForService) Save(ctx context.Context, i *entity.Insumo) error   { return nil }
func (f *fakeInsumoRepoForService) Update(ctx context.Context, i *entity.Insumo) error { return nil }
func (f *fakeInsumoRepoForService) Delete(ctx context.Context, id, userID string) error { return nil }
func (f *fakeInsumoRepoForService) SaveBatch(ctx context.Context, i []*entity.Insumo) error { return nil }
func (f *fakeInsumoRepoForService) AdjustStock(ctx context.Context, insumoID, userID string, delta float64) error {
	return nil
}
func (f *fakeInsumoRepoForService) InsertInsumoMovement(ctx context.Context, m *entity.InsumoMovement) error {
	return nil
}
func (f *fakeInsumoRepoForService) AdjustLocationInsumoStock(ctx context.Context, locationID, insumoID string, delta float64) error {
	return nil
}
func (f *fakeInsumoRepoForService) GetLocationInsumoStock(ctx context.Context, locationID, insumoID string) (float64, error) {
	return 0, nil
}

func (f *fakeInsumoRepoForService) GetByID(ctx context.Context, id, userID string) (*entity.Insumo, error) {
	return f.insumos[id], nil
}

func (f *fakeInsumoRepoForService) ListByUserID(ctx context.Context, userID string) ([]*entity.Insumo, error) {
	var out []*entity.Insumo
	for _, i := range f.insumos {
		out = append(out, i)
	}
	return out, nil
}

func stockPtr(v float64) *float64 { return &v }

func TestSearchServices_FiltersByNameCaseInsensitive(t *testing.T) {
	repo := &fakeServiceRepo{services: []*entity.Service{
		{ID: "s1", UserID: "u1", Name: "Impresión 3D Llaveros"},
		{ID: "s2", UserID: "u1", Name: "Grabado láser"},
	}}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	results, err := uc.SearchServices(context.Background(), "u1", "llavero")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "s1", results[0].ID)
	assert.Equal(t, "Impresión 3D Llaveros", results[0].Name)
}

// TestSearchServices_AccentInsensitive reproduces a live bug: a customer
// asked for "soporte para laptop" and the agent, before quoting, searched
// "impresión 3D" (natural accented Spanish) against a service literally
// named "Impresion 3D" (no accent, as stored in the real catalog) — a plain
// strings.Contains found zero results and the agent wrongly told the
// customer the service didn't exist.
func TestSearchServices_AccentInsensitive(t *testing.T) {
	repo := &fakeServiceRepo{services: []*entity.Service{
		{ID: "s1", UserID: "u1", Name: "Impresion 3D"},
	}}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	results, err := uc.SearchServices(context.Background(), "u1", "impresión 3D")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "s1", results[0].ID)
}

func TestSearchServices_EmptyQueryReturnsAll(t *testing.T) {
	repo := &fakeServiceRepo{services: []*entity.Service{
		{ID: "s1", UserID: "u1", Name: "A"},
		{ID: "s2", UserID: "u1", Name: "B"},
	}}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	results, err := uc.SearchServices(context.Background(), "u1", "")

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestCheckFeasibility_ServiceNotOffered(t *testing.T) {
	repo := &fakeServiceRepo{}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	offered, feasible, err := uc.CheckFeasibility(context.Background(), "u1", "nope", 50)

	assert.NoError(t, err)
	assert.False(t, offered)
	assert.False(t, feasible)
}

func TestCheckFeasibility_NoRecipeIsAlwaysFeasible(t *testing.T) {
	repo := &fakeServiceRepo{
		services: []*entity.Service{{ID: "s1", UserID: "u1", Name: "Llaveros"}},
	}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	offered, feasible, err := uc.CheckFeasibility(context.Background(), "u1", "s1", 50)

	assert.NoError(t, err)
	assert.True(t, offered)
	assert.True(t, feasible)
}

func TestCheckFeasibility_EnoughInsumoStock(t *testing.T) {
	repo := &fakeServiceRepo{
		services: []*entity.Service{{ID: "s1", UserID: "u1", Name: "Llaveros"}},
		insumosByID: map[string][]entity.ServiceInsumoDetail{
			"s1": {{ServiceInsumo: entity.ServiceInsumo{InsumoID: "ins1", QuantityPerUnit: 0.02}}},
		},
	}
	insumoRepo := &fakeInsumoRepoForService{insumos: map[string]*entity.Insumo{
		"ins1": {ID: "ins1", UserID: "u1", Name: "Filamento negro", Stock: stockPtr(5.0)},
	}}
	uc := usecase.NewServiceUsecases(repo, nil, insumoRepo)

	// 50 unidades * 0.02kg = 1kg necesario, hay 5kg de stock.
	offered, feasible, err := uc.CheckFeasibility(context.Background(), "u1", "s1", 50)

	assert.NoError(t, err)
	assert.True(t, offered)
	assert.True(t, feasible)
}

func TestCheckFeasibility_NotEnoughInsumoStock(t *testing.T) {
	repo := &fakeServiceRepo{
		services: []*entity.Service{{ID: "s1", UserID: "u1", Name: "Llaveros"}},
		insumosByID: map[string][]entity.ServiceInsumoDetail{
			"s1": {{ServiceInsumo: entity.ServiceInsumo{InsumoID: "ins1", QuantityPerUnit: 0.5}}},
		},
	}
	insumoRepo := &fakeInsumoRepoForService{insumos: map[string]*entity.Insumo{
		"ins1": {ID: "ins1", UserID: "u1", Name: "Filamento negro", Stock: stockPtr(5.0)},
	}}
	uc := usecase.NewServiceUsecases(repo, nil, insumoRepo)

	// 50 unidades * 0.5kg = 25kg necesario, solo hay 5kg.
	offered, feasible, err := uc.CheckFeasibility(context.Background(), "u1", "s1", 50)

	assert.NoError(t, err)
	assert.True(t, offered)
	assert.False(t, feasible)
}

func TestCheckFeasibility_NilStockNeverBlocks(t *testing.T) {
	repo := &fakeServiceRepo{
		services: []*entity.Service{{ID: "s1", UserID: "u1", Name: "Llaveros"}},
		insumosByID: map[string][]entity.ServiceInsumoDetail{
			"s1": {{ServiceInsumo: entity.ServiceInsumo{InsumoID: "ins1", QuantityPerUnit: 0.5}}},
		},
	}
	insumoRepo := &fakeInsumoRepoForService{insumos: map[string]*entity.Insumo{
		"ins1": {ID: "ins1", UserID: "u1", Name: "Filamento negro", Stock: nil},
	}}
	uc := usecase.NewServiceUsecases(repo, nil, insumoRepo)

	offered, feasible, err := uc.CheckFeasibility(context.Background(), "u1", "s1", 50)

	assert.NoError(t, err)
	assert.True(t, offered)
	assert.True(t, feasible)
}

func TestGetServiceVariables_ReturnsCustomerSafeSubset(t *testing.T) {
	unitCost := 500.0
	repo := &fakeServiceRepo{services: []*entity.Service{{
		ID: "s1", UserID: "u1", Name: "Llaveros 3D",
		VariablesSchema: []catalog.VariableDefinition{
			{
				Name: "color", Label: "Color", Type: catalog.VariableTypeSelect, Required: true,
				Options: []catalog.VariableOption{
					{Label: "Negro", Value: "negro", UnitCost: 0},
					{Label: "Blanco", Value: "blanco", UnitCost: 50},
				},
			},
			{Name: "cantidad", Label: "Cantidad", Type: catalog.VariableTypeNumber, Unit: "u", UnitCost: &unitCost, Required: true},
		},
	}}}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	results, err := uc.GetServiceVariables(context.Background(), "u1", "s1")

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "color", results[0].Name)
	assert.Equal(t, "select", results[0].Type)
	assert.True(t, results[0].Required)
	assert.Equal(t, []dto.ServiceVariableOption{
		{Label: "Negro", Value: "negro"},
		{Label: "Blanco", Value: "blanco"},
	}, results[0].Options)
	assert.Equal(t, "cantidad", results[1].Name)
	assert.Equal(t, "number", results[1].Type)
	assert.Equal(t, "u", results[1].Unit)
}

func TestGetServiceVariables_ServiceNotFound(t *testing.T) {
	repo := &fakeServiceRepo{}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	_, err := uc.GetServiceVariables(context.Background(), "u1", "nope")

	assert.ErrorIs(t, err, usecase.ErrServiceNotFoundForPricing)
}

func TestGetServicePrice_ComputesRealFormula(t *testing.T) {
	repo := &fakeServiceRepo{services: []*entity.Service{{
		ID: "s1", UserID: "u1", Name: "Llaveros 3D", Formula: "cantidad * 100 + color",
		VariablesSchema: []catalog.VariableDefinition{
			{
				Name: "color", Type: catalog.VariableTypeSelect, Required: true,
				Options: []catalog.VariableOption{{Value: "negro", UnitCost: 200}},
			},
			{Name: "cantidad", Type: catalog.VariableTypeNumber, Required: true},
		},
	}}}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	price, err := uc.GetServicePrice(context.Background(), "u1", "s1", map[string]any{
		"color":    "negro",
		"cantidad": 50.0,
	})

	assert.NoError(t, err)
	assert.Equal(t, 5200.0, price) // 50*100 + 200
}

func TestGetServicePrice_MissingRequiredVariable(t *testing.T) {
	repo := &fakeServiceRepo{services: []*entity.Service{{
		ID: "s1", UserID: "u1", Formula: "cantidad",
		VariablesSchema: []catalog.VariableDefinition{
			{Name: "cantidad", Type: catalog.VariableTypeNumber, Required: true},
		},
	}}}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	_, err := uc.GetServicePrice(context.Background(), "u1", "s1", map[string]any{})

	assert.ErrorIs(t, err, usecase.ErrMissingRequiredVariable)
}

func TestGetServicePrice_UnknownVariableName(t *testing.T) {
	repo := &fakeServiceRepo{services: []*entity.Service{{
		ID: "s1", UserID: "u1", Formula: "1",
		VariablesSchema: []catalog.VariableDefinition{},
	}}}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	_, err := uc.GetServicePrice(context.Background(), "u1", "s1", map[string]any{"inventada": 1.0})

	assert.ErrorIs(t, err, usecase.ErrUnknownVariableName)
}

func TestGetServicePrice_InvalidValueType(t *testing.T) {
	repo := &fakeServiceRepo{services: []*entity.Service{{
		ID: "s1", UserID: "u1", Formula: "cantidad",
		VariablesSchema: []catalog.VariableDefinition{
			{Name: "cantidad", Type: catalog.VariableTypeNumber, Required: true},
		},
	}}}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	_, err := uc.GetServicePrice(context.Background(), "u1", "s1", map[string]any{"cantidad": "no-es-numero"})

	assert.ErrorIs(t, err, usecase.ErrInvalidVariableValue)
}

func TestGetServicePrice_ServiceNotFound(t *testing.T) {
	repo := &fakeServiceRepo{}
	uc := usecase.NewServiceUsecases(repo, nil, &fakeInsumoRepoForService{})

	_, err := uc.GetServicePrice(context.Background(), "u1", "nope", map[string]any{})

	assert.ErrorIs(t, err, usecase.ErrServiceNotFoundForPricing)
}
