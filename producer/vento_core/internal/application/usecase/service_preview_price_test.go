package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

func TestPreviewPrice_ValidVariables(t *testing.T) {
	serviceRepo := newFakeServiceRepository()
	svc := printPrintingService()
	serviceRepo.services[svc.ID] = svc

	uc := NewServiceUsecases(serviceRepo, nil, nil)

	vars := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
		{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
	}

	unitPrice, err := uc.PreviewPrice(context.Background(), "user-1", svc.ID, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if unitPrice != 300 {
		t.Errorf("expected unit_price=300, got %v", unitPrice)
	}
}

func TestPreviewPrice_InvalidOptionSurfacesError(t *testing.T) {
	serviceRepo := newFakeServiceRepository()
	svc := printPrintingService()
	serviceRepo.services[svc.ID] = svc

	uc := NewServiceUsecases(serviceRepo, nil, nil)

	vars := []entity.OrderItemVariable{
		{Name: "material", Type: "select", OptionValue: ptrStr("PVC")},
		{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
	}

	_, err := uc.PreviewPrice(context.Background(), "user-1", svc.ID, vars)
	if err == nil {
		t.Fatal("expected error for invalid select option")
	}
	if !errors.Is(err, ErrInvalidVariableOption) {
		t.Errorf("expected ErrInvalidVariableOption, got %v", err)
	}
}

func TestPreviewPrice_ServiceNotFound(t *testing.T) {
	serviceRepo := newFakeServiceRepository()
	uc := NewServiceUsecases(serviceRepo, nil, nil)

	_, err := uc.PreviewPrice(context.Background(), "user-1", "missing-svc", nil)
	if err == nil {
		t.Fatal("expected error for missing service")
	}
}
