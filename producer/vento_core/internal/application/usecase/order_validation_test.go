package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

func newMinimalUC() *OrderUsecases {
	return NewOrderUsecases(
		newFakeOrderRepository(),
		newFakeServiceRepository(),
		newFakeProductRepository(),
		newFakeInsumoRepository(),
		newFakeLocationRepository(),
	)
}

func productItem(quantity int) entity.OrderItem {
	return entity.OrderItem{
		ProductID: "prod-1",
		Type:      entity.OrderItemTypeProduct,
		Name:      "Widget",
		Quantity:  quantity,
		UnitPrice: 100,
	}
}

func TestCreateOrder_EmptyItems_ReturnsErrInvalidOrderItems(t *testing.T) {
	uc := newMinimalUC()
	req := dto.CreateOrderRequest{
		ClientName: "Ana",
		Status:     "pending",
		Items:      []entity.OrderItem{},
	}
	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if !errors.Is(err, ErrInvalidOrderItems) {
		t.Fatalf("expected ErrInvalidOrderItems, got %v", err)
	}
}

func TestCreateOrder_ZeroQuantityItem_ReturnsErrInvalidOrderItems(t *testing.T) {
	uc := newMinimalUC()
	req := dto.CreateOrderRequest{
		ClientName: "Ana",
		Status:     "pending",
		Items:      []entity.OrderItem{productItem(0)},
	}
	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if !errors.Is(err, ErrInvalidOrderItems) {
		t.Fatalf("expected ErrInvalidOrderItems, got %v", err)
	}
}

func TestCreateOrder_NegativeQuantityItem_ReturnsErrInvalidOrderItems(t *testing.T) {
	uc := newMinimalUC()
	req := dto.CreateOrderRequest{
		ClientName: "Ana",
		Status:     "pending",
		Items:      []entity.OrderItem{productItem(-1)},
	}
	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if !errors.Is(err, ErrInvalidOrderItems) {
		t.Fatalf("expected ErrInvalidOrderItems, got %v", err)
	}
}

func TestUpdateOrder_EmptyItems_ReturnsErrInvalidOrderItems(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	uc := NewOrderUsecases(orderRepo, newFakeServiceRepository(), newFakeProductRepository(), newFakeInsumoRepository(), newFakeLocationRepository())

	// Pre-create a valid order
	existing := &entity.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: entity.StatusPending,
		Items:  []entity.OrderItem{productItem(2)},
	}
	orderRepo.orders[existing.ID] = existing

	empty := []entity.OrderItem{}
	req := dto.UpdateOrderRequest{Items: &empty}
	_, err := uc.UpdateOrder(context.Background(), "user-1", "order-1", req)
	if !errors.Is(err, ErrInvalidOrderItems) {
		t.Fatalf("expected ErrInvalidOrderItems, got %v", err)
	}
}

func TestUpdateOrder_ZeroQuantityItem_ReturnsErrInvalidOrderItems(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	uc := NewOrderUsecases(orderRepo, newFakeServiceRepository(), newFakeProductRepository(), newFakeInsumoRepository(), newFakeLocationRepository())

	existing := &entity.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: entity.StatusPending,
		Items:  []entity.OrderItem{productItem(2)},
	}
	orderRepo.orders[existing.ID] = existing

	badItems := []entity.OrderItem{productItem(0)}
	req := dto.UpdateOrderRequest{Items: &badItems}
	_, err := uc.UpdateOrder(context.Background(), "user-1", "order-1", req)
	if !errors.Is(err, ErrInvalidOrderItems) {
		t.Fatalf("expected ErrInvalidOrderItems, got %v", err)
	}
}

func TestUpdateOrder_ItemsLockedByPayment_ReturnsErrOrderLockedByPayment(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	uc := NewOrderUsecases(orderRepo, newFakeServiceRepository(), newFakeProductRepository(), newFakeInsumoRepository(), newFakeLocationRepository())

	existing := &entity.Order{
		ID:     "order-2",
		UserID: "user-1",
		Status: entity.StatusConfirmed,
		Items:  []entity.OrderItem{productItem(1)},
	}
	orderRepo.orders[existing.ID] = existing
	// Register a payment so the order is locked
	orderRepo.payments[existing.ID] = []entity.OrderPayment{{
		OrderID: existing.ID,
		Amount:  100,
	}}

	newItems := []entity.OrderItem{productItem(3)}
	req := dto.UpdateOrderRequest{Items: &newItems}
	_, err := uc.UpdateOrder(context.Background(), "user-1", "order-2", req)
	if !errors.Is(err, ErrOrderLockedByPayment) {
		t.Fatalf("expected ErrOrderLockedByPayment, got %v", err)
	}
}
