package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// TestCreateOrder_RejectsLocationNotOwnedByUser guards against a tenant
// isolation hole: a client could pass another tenant's location UUID in
// POST /api/v1/orders and mutate that tenant's location_stock/movements
// ledger. The order's location_id must belong to the authenticated user.
func TestCreateOrder_RejectsLocationNotOwnedByUser(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	locationRepo := newFakeLocationRepository()
	// location belongs to a different tenant
	otherTenantLocation := entity.NewLocation("user-2", "Sucursal Ajena", "", "", true)
	locationRepo.locations[otherTenantLocation.ID] = otherTenantLocation

	productRepo.products["prod-1"] = productWithStock("prod-1", 10)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, locationRepo)

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		LocationID: &otherTenantLocation.ID,
		Items: []entity.OrderItem{
			{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 3, UnitPrice: 10},
		},
	}

	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err == nil {
		t.Fatal("expected error for location not owned by user")
	}
	if !errors.Is(err, ErrLocationNotOwnedByUser) {
		t.Errorf("expected ErrLocationNotOwnedByUser, got %v", err)
	}

	if len(orderRepo.orders) != 0 {
		t.Errorf("expected no order to be saved, got %d", len(orderRepo.orders))
	}
	if productRepo.products["prod-1"].Stock != 10 {
		t.Errorf("expected stock unchanged at 10, got %v", productRepo.products["prod-1"].Stock)
	}
}

// TestCreateOrder_AcceptsLocationOwnedByUser is the corresponding happy path:
// a location that does belong to the authenticated user must be accepted.
func TestCreateOrder_AcceptsLocationOwnedByUser(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	locationRepo := newFakeLocationRepository()
	myLocation := entity.NewLocation("user-1", "Local Central", "", "", true)
	locationRepo.locations[myLocation.ID] = myLocation

	productRepo.products["prod-1"] = productWithStock("prod-1", 10)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, locationRepo)

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		LocationID: &myLocation.ID,
		Items: []entity.OrderItem{
			{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 3, UnitPrice: 10},
		},
	}

	res, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.LocationID != myLocation.ID {
		t.Errorf("expected LocationID=%s, got %s", myLocation.ID, res.LocationID)
	}
}
