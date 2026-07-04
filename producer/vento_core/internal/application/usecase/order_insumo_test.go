package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// ---- fakeInsumoRepository ----

type fakeInsumoRepository struct {
	insumos       map[string]*entity.Insumo
	movements     []entity.InsumoMovement
	locationStock map[string]float64
}

func newFakeInsumoRepository() *fakeInsumoRepository {
	return &fakeInsumoRepository{insumos: make(map[string]*entity.Insumo)}
}

func (f *fakeInsumoRepository) Save(ctx context.Context, insumo *entity.Insumo) error {
	f.insumos[insumo.ID] = insumo
	return nil
}
func (f *fakeInsumoRepository) Update(ctx context.Context, insumo *entity.Insumo) error {
	f.insumos[insumo.ID] = insumo
	return nil
}
func (f *fakeInsumoRepository) Delete(ctx context.Context, id string, userID string) error {
	delete(f.insumos, id)
	return nil
}
func (f *fakeInsumoRepository) GetByID(ctx context.Context, id string, userID string) (*entity.Insumo, error) {
	return f.insumos[id], nil
}
func (f *fakeInsumoRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Insumo, error) {
	var res []*entity.Insumo
	for _, i := range f.insumos {
		res = append(res, i)
	}
	return res, nil
}
func (f *fakeInsumoRepository) SaveBatch(ctx context.Context, insumos []*entity.Insumo) error {
	for _, i := range insumos {
		f.insumos[i.ID] = i
	}
	return nil
}
func (f *fakeInsumoRepository) AdjustStock(ctx context.Context, insumoID string, userID string, delta float64) error {
	i, ok := f.insumos[insumoID]
	if !ok {
		return nil
	}
	var newStock float64
	if i.Stock != nil {
		newStock = *i.Stock
	}
	newStock += delta
	i.Stock = &newStock
	return nil
}
func (f *fakeInsumoRepository) InsertInsumoMovement(ctx context.Context, movement *entity.InsumoMovement) error {
	f.movements = append(f.movements, *movement)
	return nil
}
func (f *fakeInsumoRepository) AdjustLocationInsumoStock(ctx context.Context, locationID string, insumoID string, delta float64) error {
	if locationID == "" {
		return nil
	}
	if f.locationStock == nil {
		f.locationStock = make(map[string]float64)
	}
	f.locationStock[locationID+"|"+insumoID] += delta
	return nil
}
func (f *fakeInsumoRepository) GetLocationInsumoStock(ctx context.Context, locationID string, insumoID string) (float64, error) {
	if f.locationStock == nil {
		return 0, nil
	}
	return f.locationStock[locationID+"|"+insumoID], nil
}

// insumoWithStock builds an Insumo with a non-nil stock pointer.
func insumoWithStock(id string, stock float64) *entity.Insumo {
	return &entity.Insumo{ID: id, UserID: "user-1", Name: "Insumo " + id, Stock: &stock}
}

// insumoWithoutStock builds an Insumo that was never loaded with stock data
// (Stock == nil) — must never block a sale.
func insumoWithoutStock(id string) *entity.Insumo {
	return &entity.Insumo{ID: id, UserID: "user-1", Name: "Insumo " + id, Stock: nil}
}

// servicePrintingWithRecipe registers a fixed recipe for printPrintingService
// in the given fakeServiceRepository, returning the same service so callers
// can attach it to order items.
func servicePrintingWithRecipe(serviceRepo *fakeServiceRepository, insumoID string, quantityPerUnit float64) *entity.Service {
	svc := printPrintingService()
	serviceRepo.services[svc.ID] = svc
	serviceRepo.recipes[svc.ID] = []entity.ServiceInsumoDetail{
		{
			ServiceInsumo: entity.ServiceInsumo{
				ServiceID:       svc.ID,
				InsumoID:        insumoID,
				QuantityPerUnit: quantityPerUnit,
			},
			InsumoName: "Insumo " + insumoID,
			Unit:       "g",
		},
	}
	return svc
}

func TestCreateOrder_ServiceWithRecipeConsumesInsumo(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	insumoRepo.insumos["insumo-1"] = insumoWithStock("insumo-1", 100)

	svc := servicePrintingWithRecipe(serviceRepo, "insumo-1", 1.2)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{
				ServiceID: svc.ID,
				Type:      entity.OrderItemTypeService,
				Name:      "Impresión 3D",
				Quantity:  2,
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

	// 1.2 per unit * 2 units = 2.4 consumed
	if *insumoRepo.insumos["insumo-1"].Stock != 97.6 {
		t.Errorf("expected stock=97.6, got %v", *insumoRepo.insumos["insumo-1"].Stock)
	}

	if len(insumoRepo.movements) != 1 {
		t.Fatalf("expected 1 movement, got %d", len(insumoRepo.movements))
	}
	m := insumoRepo.movements[0]
	if m.QuantityDelta != -2.4 {
		t.Errorf("expected QuantityDelta=-2.4, got %v", m.QuantityDelta)
	}
	if m.Reason != "production" {
		t.Errorf("expected Reason=production, got %v", m.Reason)
	}
}

func TestCreateOrder_InsufficientInsumoStockRejectsOrder(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	insumoRepo.insumos["insumo-1"] = insumoWithStock("insumo-1", 1)

	svc := servicePrintingWithRecipe(serviceRepo, "insumo-1", 1.2)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{
				ServiceID: svc.ID,
				Type:      entity.OrderItemTypeService,
				Name:      "Impresión 3D",
				Quantity:  2,
				Variables: []entity.OrderItemVariable{
					{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
					{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
				},
			},
		},
	}

	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err == nil {
		t.Fatal("expected error for insufficient insumo stock")
	}
	if !errors.Is(err, ErrInsufficientStock) {
		t.Errorf("expected ErrInsufficientStock, got %v", err)
	}

	if len(orderRepo.orders) != 0 {
		t.Errorf("expected no order to be saved, got %d", len(orderRepo.orders))
	}
	if *insumoRepo.insumos["insumo-1"].Stock != 1 {
		t.Errorf("expected stock unchanged at 1, got %v", *insumoRepo.insumos["insumo-1"].Stock)
	}
}

func TestUpdateOrderStatus_CancelledRestoresInsumoStock(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	insumoRepo.insumos["insumo-1"] = insumoWithStock("insumo-1", 100)

	svc := servicePrintingWithRecipe(serviceRepo, "insumo-1", 1.2)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{
				ServiceID: svc.ID,
				Type:      entity.OrderItemTypeService,
				Name:      "Impresión 3D",
				Quantity:  2,
				Variables: []entity.OrderItemVariable{
					{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
					{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
				},
			},
		},
	}
	res, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error creating order: %v", err)
	}
	if *insumoRepo.insumos["insumo-1"].Stock != 97.6 {
		t.Fatalf("expected stock=97.6 after create, got %v", *insumoRepo.insumos["insumo-1"].Stock)
	}

	_, err = uc.UpdateOrderStatus(context.Background(), "user-1", res.ID, "cancelled")
	if err != nil {
		t.Fatalf("unexpected error cancelling order: %v", err)
	}

	if *insumoRepo.insumos["insumo-1"].Stock != 100 {
		t.Errorf("expected stock restored to 100, got %v", *insumoRepo.insumos["insumo-1"].Stock)
	}

	last := insumoRepo.movements[len(insumoRepo.movements)-1]
	if last.Reason != "cancellation" {
		t.Errorf("expected Reason=cancellation, got %v", last.Reason)
	}
	if last.QuantityDelta != 2.4 {
		t.Errorf("expected QuantityDelta=+2.4, got %v", last.QuantityDelta)
	}
}

// TestDeleteOrder_AlreadyCancelledDoesNotDoubleCreditInsumoStock mirrors
// TestDeleteOrder_AlreadyCancelledDoesNotDoubleCreditStock (order_stock_test.go)
// but for insumos: deleting an order that was already cancelled (and
// therefore already had its insumo stock restored once) must NOT restore it
// a second time.
func TestDeleteOrder_AlreadyCancelledDoesNotDoubleCreditInsumoStock(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	insumoRepo.insumos["insumo-1"] = insumoWithStock("insumo-1", 100)

	svc := servicePrintingWithRecipe(serviceRepo, "insumo-1", 1.2)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{
				ServiceID: svc.ID,
				Type:      entity.OrderItemTypeService,
				Name:      "Impresión 3D",
				Quantity:  2,
				Variables: []entity.OrderItemVariable{
					{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
					{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
				},
			},
		},
	}
	res, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error creating order: %v", err)
	}
	if *insumoRepo.insumos["insumo-1"].Stock != 97.6 {
		t.Fatalf("expected stock=97.6 after create, got %v", *insumoRepo.insumos["insumo-1"].Stock)
	}

	if _, err := uc.UpdateOrderStatus(context.Background(), "user-1", res.ID, "cancelled"); err != nil {
		t.Fatalf("unexpected error cancelling order: %v", err)
	}
	if *insumoRepo.insumos["insumo-1"].Stock != 100 {
		t.Fatalf("expected stock restored to 100 after cancel, got %v", *insumoRepo.insumos["insumo-1"].Stock)
	}

	if err := uc.DeleteOrder(context.Background(), "user-1", res.ID); err != nil {
		t.Fatalf("unexpected error deleting order: %v", err)
	}

	if *insumoRepo.insumos["insumo-1"].Stock != 100 {
		t.Errorf("expected stock to remain 100 (no double credit), got %v", *insumoRepo.insumos["insumo-1"].Stock)
	}
}

func TestCreateOrder_InsumoWithNilStockDoesNotBlockSale(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	insumoRepo.insumos["insumo-1"] = insumoWithoutStock("insumo-1")

	svc := servicePrintingWithRecipe(serviceRepo, "insumo-1", 1.2)

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	req := dto.CreateOrderRequest{
		ClientName: "Cliente Test",
		Status:     "pending",
		Items: []entity.OrderItem{
			{
				ServiceID: svc.ID,
				Type:      entity.OrderItemTypeService,
				Name:      "Impresión 3D",
				Quantity:  2,
				Variables: []entity.OrderItemVariable{
					{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
					{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
				},
			},
		},
	}

	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("expected sale to proceed despite nil stock, got error: %v", err)
	}
}
