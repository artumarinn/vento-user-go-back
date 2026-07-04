package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

func productWithStock(id string, stock float64) *entity.Product {
	return &entity.Product{ID: id, UserID: "user-1", Name: "Product " + id, Stock: stock}
}

func TestCreateOrder_DecrementsProductStock(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	productRepo.products["prod-1"] = productWithStock("prod-1", 10)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 3, UnitPrice: 10},
		},
	}

	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if productRepo.products["prod-1"].Stock != 7 {
		t.Errorf("expected stock=7, got %v", productRepo.products["prod-1"].Stock)
	}

	if len(productRepo.movements) != 1 {
		t.Fatalf("expected 1 movement, got %d", len(productRepo.movements))
	}
	m := productRepo.movements[0]
	if m.QuantityDelta != -3 {
		t.Errorf("expected QuantityDelta=-3, got %v", m.QuantityDelta)
	}
	if m.Reason != "sale" {
		t.Errorf("expected Reason=sale, got %v", m.Reason)
	}
}

func TestCreateOrder_InsufficientStockRejectsOrder(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	productRepo.products["prod-1"] = productWithStock("prod-1", 2)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 5, UnitPrice: 10},
		},
	}

	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
	if !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}

	if len(orderRepo.orders) != 0 {
		t.Errorf("expected no order to be saved, got %d", len(orderRepo.orders))
	}
	if productRepo.products["prod-1"].Stock != 2 {
		t.Errorf("expected stock unchanged at 2, got %v", productRepo.products["prod-1"].Stock)
	}
}

func TestCreateOrder_ServiceItemsDoNotTouchStock(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	svc := printPrintingService()
	serviceRepo.services[svc.ID] = svc

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{
				ServiceID: svc.ID,
				Type:      entity.OrderItemTypeService,
				Name:      "Impresión 3D",
				Quantity:  1,
				Variables: []entity.OrderItemVariable{
					{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
					{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
				},
			},
		},
	}

	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(productRepo.movements) != 0 {
		t.Errorf("expected no stock movements for service items, got %d", len(productRepo.movements))
	}
}

func TestUpdateOrderStatus_CancelledRestoresStock(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	productRepo.products["prod-1"] = productWithStock("prod-1", 10)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 3, UnitPrice: 10},
		},
	}
	res, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error creating order: %v", err)
	}
	if productRepo.products["prod-1"].Stock != 7 {
		t.Fatalf("expected stock=7 after create, got %v", productRepo.products["prod-1"].Stock)
	}

	_, err = uc.UpdateOrderStatus(context.Background(), "user-1", res.ID, "cancelled")
	if err != nil {
		t.Fatalf("unexpected error cancelling order: %v", err)
	}

	if productRepo.products["prod-1"].Stock != 10 {
		t.Errorf("expected stock restored to 10, got %v", productRepo.products["prod-1"].Stock)
	}

	last := productRepo.movements[len(productRepo.movements)-1]
	if last.Reason != "cancellation" {
		t.Errorf("expected Reason=cancellation, got %v", last.Reason)
	}
	if last.QuantityDelta != 3 {
		t.Errorf("expected QuantityDelta=+3, got %v", last.QuantityDelta)
	}
}

func TestDeleteOrder_RestoresStockBeforeDeleting(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	productRepo.products["prod-1"] = productWithStock("prod-1", 10)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 3, UnitPrice: 10},
		},
	}
	res, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error creating order: %v", err)
	}
	if productRepo.products["prod-1"].Stock != 7 {
		t.Fatalf("expected stock=7 after create, got %v", productRepo.products["prod-1"].Stock)
	}

	if err := uc.DeleteOrder(context.Background(), "user-1", res.ID); err != nil {
		t.Fatalf("unexpected error deleting order: %v", err)
	}

	if productRepo.products["prod-1"].Stock != 10 {
		t.Errorf("expected stock restored to 10, got %v", productRepo.products["prod-1"].Stock)
	}

	last := productRepo.movements[len(productRepo.movements)-1]
	if last.Reason != "cancellation" {
		t.Errorf("expected Reason=cancellation, got %v", last.Reason)
	}
	if last.QuantityDelta != 3 {
		t.Errorf("expected QuantityDelta=+3, got %v", last.QuantityDelta)
	}

	if _, ok := orderRepo.orders[res.ID]; ok {
		t.Errorf("expected order to be deleted from repo")
	}
}

// TestDeleteOrder_AlreadyCancelledDoesNotDoubleCreditStock guards against a
// real bug found via live testing: deleting an order that was already
// cancelled (and therefore already had its stock restored once) must NOT
// restore it a second time.
func TestDeleteOrder_AlreadyCancelledDoesNotDoubleCreditStock(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	productRepo.products["prod-1"] = productWithStock("prod-1", 10)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 3, UnitPrice: 10},
		},
	}
	res, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error creating order: %v", err)
	}
	if productRepo.products["prod-1"].Stock != 7 {
		t.Fatalf("expected stock=7 after create, got %v", productRepo.products["prod-1"].Stock)
	}

	if _, err := uc.UpdateOrderStatus(context.Background(), "user-1", res.ID, "cancelled"); err != nil {
		t.Fatalf("unexpected error cancelling order: %v", err)
	}
	if productRepo.products["prod-1"].Stock != 10 {
		t.Fatalf("expected stock restored to 10 after cancel, got %v", productRepo.products["prod-1"].Stock)
	}

	if err := uc.DeleteOrder(context.Background(), "user-1", res.ID); err != nil {
		t.Fatalf("unexpected error deleting order: %v", err)
	}

	if productRepo.products["prod-1"].Stock != 10 {
		t.Errorf("expected stock to remain 10 (no double credit), got %v", productRepo.products["prod-1"].Stock)
	}
}

func TestUpdateOrder_AdjustsStockByDelta(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	productRepo.products["prod-1"] = productWithStock("prod-1", 10)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 3, UnitPrice: 10},
		},
	}
	res, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error creating order: %v", err)
	}
	if productRepo.products["prod-1"].Stock != 7 {
		t.Fatalf("expected stock=7 after create, got %v", productRepo.products["prod-1"].Stock)
	}

	newItems := []entity.OrderItem{
		{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 5, UnitPrice: 10},
	}
	updateReq := dto.UpdateOrderRequest{Items: &newItems}

	_, err = uc.UpdateOrder(context.Background(), "user-1", res.ID, updateReq)
	if err != nil {
		t.Fatalf("unexpected error updating order: %v", err)
	}

	if productRepo.products["prod-1"].Stock != 5 {
		t.Errorf("expected stock=5 after increasing quantity to 5 (delta -2 more), got %v", productRepo.products["prod-1"].Stock)
	}

	last := productRepo.movements[len(productRepo.movements)-1]
	if last.Reason != "adjustment" {
		t.Errorf("expected Reason=adjustment, got %v", last.Reason)
	}
	if last.QuantityDelta != -2 {
		t.Errorf("expected QuantityDelta=-2, got %v", last.QuantityDelta)
	}
}

func TestUpdateOrder_InsufficientStockOnIncreaseRejects(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	productRepo.products["prod-1"] = productWithStock("prod-1", 4)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 3, UnitPrice: 10},
		},
	}
	res, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error creating order: %v", err)
	}
	// stock now 4-3=1, available for increase is only 1 unit more (cannot reach +20)
	newItems := []entity.OrderItem{
		{ProductID: "prod-1", Type: entity.OrderItemTypeProduct, Name: "Product prod-1", Quantity: 23, UnitPrice: 10},
	}
	updateReq := dto.UpdateOrderRequest{Items: &newItems}

	_, err = uc.UpdateOrder(context.Background(), "user-1", res.ID, updateReq)
	if err == nil {
		t.Fatal("expected error for insufficient stock on increase")
	}
	if !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}

	stored := orderRepo.orders[res.ID]
	if len(stored.Items) != 1 || stored.Items[0].Quantity != 3 {
		t.Errorf("expected order to remain unmodified with quantity=3, got %+v", stored.Items)
	}
}
