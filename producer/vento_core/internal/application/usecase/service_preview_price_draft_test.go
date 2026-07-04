package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

func TestPreviewPriceDraft_ValidVariables(t *testing.T) {
	svc := printPrintingService()

	uc := NewServiceUsecases(nil, nil)

	vars := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
		{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
	}

	draftPrice, err := uc.PreviewPriceDraft(svc.Formula, svc.VariablesSchema, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draftPrice != 300 {
		t.Errorf("expected unit_price=300, got %v", draftPrice)
	}

	// Must match what PreviewPrice would compute for an equivalent saved service.
	serviceRepo := newFakeServiceRepository()
	serviceRepo.services[svc.ID] = svc
	savedUC := NewServiceUsecases(serviceRepo, nil)
	savedPrice, err := savedUC.PreviewPrice(context.Background(), "user-1", svc.ID, vars)
	if err != nil {
		t.Fatalf("unexpected error from PreviewPrice: %v", err)
	}
	if draftPrice != savedPrice {
		t.Errorf("expected draft price to match saved-service preview price: draft=%v saved=%v", draftPrice, savedPrice)
	}
}

func TestPreviewPriceDraft_UnknownVariableInFormula(t *testing.T) {
	uc := NewServiceUsecases(nil, nil)

	formula := "material * peso * tiempo"
	schema := []catalog.VariableDefinition{
		{
			Name:     "material",
			Type:     catalog.VariableTypeSelect,
			Required: true,
			Options: []catalog.VariableOption{
				{Label: "PLA", Value: "PLA", UnitCost: 2},
			},
		},
		{
			Name:     "peso",
			Type:     catalog.VariableTypeNumber,
			Required: true,
		},
	}

	vars := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
		{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
	}

	_, err := uc.PreviewPriceDraft(formula, schema, vars)
	if err == nil {
		t.Fatal("expected error for formula referencing undeclared identifier 'tiempo'")
	}
}

func TestPreviewPriceDraft_InvalidOptionSurfacesError(t *testing.T) {
	svc := printPrintingService()
	uc := NewServiceUsecases(nil, nil)

	vars := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("PVC")},
		{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
	}

	_, err := uc.PreviewPriceDraft(svc.Formula, svc.VariablesSchema, vars)
	if err == nil {
		t.Fatal("expected error for invalid select option")
	}
	if !errors.Is(err, ErrInvalidVariableOption) {
		t.Errorf("expected ErrInvalidVariableOption, got %v", err)
	}
}
