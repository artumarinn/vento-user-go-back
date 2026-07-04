package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// ---- fakeOrderRepository ----

type fakeOrderRepository struct {
	orders   map[string]*entity.Order
	payments map[string][]entity.OrderPayment
}

func newFakeOrderRepository() *fakeOrderRepository {
	return &fakeOrderRepository{
		orders:   make(map[string]*entity.Order),
		payments: make(map[string][]entity.OrderPayment),
	}
}

func (f *fakeOrderRepository) Save(ctx context.Context, order *entity.Order) error {
	f.orders[order.ID] = order
	return nil
}
func (f *fakeOrderRepository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	return f.orders[id], nil
}
func (f *fakeOrderRepository) GetByIDForUser(ctx context.Context, id string, userID string) (*entity.Order, error) {
	o, ok := f.orders[id]
	if !ok {
		return nil, nil
	}
	return o, nil
}
func (f *fakeOrderRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Order, error) {
	var res []*entity.Order
	for _, o := range f.orders {
		res = append(res, o)
	}
	return res, nil
}
func (f *fakeOrderRepository) ListByUserIDAndLocation(ctx context.Context, userID string, locationID string) ([]*entity.Order, error) {
	var res []*entity.Order
	for _, o := range f.orders {
		if o.LocationID == locationID {
			res = append(res, o)
		}
	}
	return res, nil
}
func (f *fakeOrderRepository) UpdateStatus(ctx context.Context, id string, userID string, status entity.OrderStatus) error {
	if o, ok := f.orders[id]; ok {
		o.Status = status
	}
	return nil
}
func (f *fakeOrderRepository) SaveBatch(ctx context.Context, orders []*entity.Order) error {
	for _, o := range orders {
		f.orders[o.ID] = o
	}
	return nil
}
func (f *fakeOrderRepository) Update(ctx context.Context, order *entity.Order) error {
	f.orders[order.ID] = order
	return nil
}
func (f *fakeOrderRepository) Delete(ctx context.Context, id, userID string) error {
	delete(f.orders, id)
	return nil
}
func (f *fakeOrderRepository) AppendStatusHistory(ctx context.Context, orderID string, status entity.OrderStatus) error {
	return nil
}
func (f *fakeOrderRepository) GetStatusHistory(ctx context.Context, orderID string) ([]entity.OrderStatusEvent, error) {
	return nil, nil
}
func (f *fakeOrderRepository) RegisterPayment(ctx context.Context, orderID string, method string, amount float64) error {
	return nil
}
func (f *fakeOrderRepository) InsertOrderPayment(ctx context.Context, payment *entity.OrderPayment) error {
	f.payments[payment.OrderID] = append(f.payments[payment.OrderID], *payment)
	return nil
}
func (f *fakeOrderRepository) ListOrderPayments(ctx context.Context, orderID string) ([]entity.OrderPayment, error) {
	return f.payments[orderID], nil
}
func (f *fakeOrderRepository) ListOrderPaymentsByUserID(ctx context.Context, userID string) ([]entity.OrderPaymentDetail, error) {
	return nil, nil
}
func (f *fakeOrderRepository) SumPaidOrderPayments(ctx context.Context, orderID string) (float64, error) {
	var total float64
	for _, p := range f.payments[orderID] {
		total += p.Amount
	}
	return total, nil
}
func (f *fakeOrderRepository) UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus string) error {
	if o, ok := f.orders[orderID]; ok {
		o.PaymentStatus = paymentStatus
	}
	return nil
}

// ---- fakeServiceRepository ----

type fakeServiceRepository struct {
	services map[string]*entity.Service
	recipes  map[string][]entity.ServiceInsumoDetail
}

func newFakeServiceRepository() *fakeServiceRepository {
	return &fakeServiceRepository{
		services: make(map[string]*entity.Service),
		recipes:  make(map[string][]entity.ServiceInsumoDetail),
	}
}

func (f *fakeServiceRepository) Save(ctx context.Context, service *entity.Service) error {
	f.services[service.ID] = service
	return nil
}
func (f *fakeServiceRepository) Update(ctx context.Context, service *entity.Service) error {
	f.services[service.ID] = service
	return nil
}
func (f *fakeServiceRepository) Delete(ctx context.Context, id string, userID string) error {
	delete(f.services, id)
	return nil
}
func (f *fakeServiceRepository) GetByID(ctx context.Context, id string, userID string) (*entity.Service, error) {
	return f.services[id], nil
}
func (f *fakeServiceRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Service, error) {
	var res []*entity.Service
	for _, s := range f.services {
		res = append(res, s)
	}
	return res, nil
}
func (f *fakeServiceRepository) SaveBatch(ctx context.Context, services []*entity.Service) error {
	for _, s := range services {
		f.services[s.ID] = s
	}
	return nil
}
func (f *fakeServiceRepository) SetServiceInsumos(ctx context.Context, serviceID string, insumos []entity.ServiceInsumo) error {
	details := make([]entity.ServiceInsumoDetail, len(insumos))
	for i, si := range insumos {
		details[i] = entity.ServiceInsumoDetail{ServiceInsumo: si}
	}
	f.recipes[serviceID] = details
	return nil
}
func (f *fakeServiceRepository) ListServiceInsumos(ctx context.Context, serviceID string) ([]entity.ServiceInsumoDetail, error) {
	return f.recipes[serviceID], nil
}

// ---- fakeProductRepository ----

type fakeProductRepository struct {
	products      map[string]*entity.Product
	movements     []entity.StockMovement
	locationStock map[string]float64
}

func newFakeProductRepository() *fakeProductRepository {
	return &fakeProductRepository{products: make(map[string]*entity.Product)}
}

func (f *fakeProductRepository) Save(ctx context.Context, product *entity.Product) error {
	f.products[product.ID] = product
	return nil
}
func (f *fakeProductRepository) Update(ctx context.Context, product *entity.Product) error {
	f.products[product.ID] = product
	return nil
}
func (f *fakeProductRepository) Delete(ctx context.Context, id string, userID string) error {
	delete(f.products, id)
	return nil
}
func (f *fakeProductRepository) GetByID(ctx context.Context, id string, userID string) (*entity.Product, error) {
	return f.products[id], nil
}
func (f *fakeProductRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Product, error) {
	var res []*entity.Product
	for _, p := range f.products {
		res = append(res, p)
	}
	return res, nil
}
func (f *fakeProductRepository) SaveBatch(ctx context.Context, products []*entity.Product) error {
	for _, p := range products {
		f.products[p.ID] = p
	}
	return nil
}
func (f *fakeProductRepository) Search(ctx context.Context, userID string, query string, category string, limit int) ([]*entity.Product, error) {
	return nil, nil
}
func (f *fakeProductRepository) AdjustStock(ctx context.Context, productID string, userID string, delta float64) error {
	p, ok := f.products[productID]
	if !ok {
		return nil
	}
	p.Stock += delta
	return nil
}
func (f *fakeProductRepository) InsertStockMovement(ctx context.Context, movement *entity.StockMovement) error {
	f.movements = append(f.movements, *movement)
	return nil
}
func (f *fakeProductRepository) AdjustLocationStock(ctx context.Context, locationID string, productID string, delta float64) error {
	if locationID == "" {
		return nil
	}
	if f.locationStock == nil {
		f.locationStock = make(map[string]float64)
	}
	f.locationStock[locationID+"|"+productID] += delta
	return nil
}
func (f *fakeProductRepository) GetLocationStock(ctx context.Context, locationID string, productID string) (float64, error) {
	if f.locationStock == nil {
		return 0, nil
	}
	return f.locationStock[locationID+"|"+productID], nil
}

func printPrintingService() *entity.Service {
	s := entity.Service(catalog.Service{
		ID:      "svc-print",
		UserID:  "user-1",
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
			{
				Name:     "peso",
				Type:     catalog.VariableTypeNumber,
				Required: true,
			},
		},
	})
	return &s
}

func TestCreateOrder_ComputesServiceItemUnitPrice(t *testing.T) {
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
				UnitPrice: 99999, // client-supplied price must be ignored
				Variables: []entity.OrderItemVariable{
					{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
					{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
				},
			},
		},
	}

	res, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(res.Items))
	}
	if res.Items[0].UnitPrice != 300 {
		t.Errorf("expected computed unit_price=300, got %v", res.Items[0].UnitPrice)
	}
	if res.TotalAmount != 300 {
		t.Errorf("expected total=300, got %v", res.TotalAmount)
	}
}

func TestCreateOrder_RejectsServiceItemWithEvaluatorError(t *testing.T) {
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
					// missing required "peso"
				},
			},
		},
	}

	_, err := uc.CreateOrder(context.Background(), "user-1", req)
	if err == nil {
		t.Fatal("expected error for incomplete service item variables")
	}
	if !errors.Is(err, ErrMissingRequiredVariable) {
		t.Errorf("expected ErrMissingRequiredVariable, got %v", err)
	}
}

func TestUpdateOrder_LockedAfterPaymentExists(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	svc := printPrintingService()
	serviceRepo.services[svc.ID] = svc

	existing := &entity.Order{
		ID:     "order-1",
		UserID: "user-1",
		Items: []entity.OrderItem{
			{ServiceID: svc.ID, Type: entity.OrderItemTypeService, Quantity: 1, UnitPrice: 300},
		},
		Total: 300,
	}
	orderRepo.orders[existing.ID] = existing
	orderRepo.payments[existing.ID] = []entity.OrderPayment{
		{OrderID: existing.ID, Amount: 100},
	}

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	newItems := []entity.OrderItem{
		{
			ServiceID: svc.ID,
			Type:      entity.OrderItemTypeService,
			Quantity:  2,
			Variables: []entity.OrderItemVariable{
				{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
				{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
			},
		},
	}
	req := dto.UpdateOrderRequest{Items: &newItems}

	_, err := uc.UpdateOrder(context.Background(), "user-1", "order-1", req)
	if err == nil {
		t.Fatal("expected ErrOrderLockedByPayment")
	}
	if !errors.Is(err, ErrOrderLockedByPayment) {
		t.Errorf("expected ErrOrderLockedByPayment, got %v", err)
	}

	stored := orderRepo.orders["order-1"]
	if stored.Total != 300 {
		t.Errorf("expected Total untouched at 300, got %v", stored.Total)
	}
}

func TestUpdateOrder_RecomputesPriceWhenUnlocked(t *testing.T) {
	orderRepo := newFakeOrderRepository()
	serviceRepo := newFakeServiceRepository()
	productRepo := newFakeProductRepository()
	insumoRepo := newFakeInsumoRepository()
	svc := printPrintingService()
	serviceRepo.services[svc.ID] = svc

	existing := &entity.Order{
		ID:     "order-2",
		UserID: "user-1",
		Items: []entity.OrderItem{
			{ServiceID: svc.ID, Type: entity.OrderItemTypeService, Quantity: 1, UnitPrice: 100},
		},
		Total: 100,
	}
	orderRepo.orders[existing.ID] = existing
	// no payments

	uc := NewOrderUsecases(orderRepo, serviceRepo, productRepo, insumoRepo, newFakeLocationRepository())

	newItems := []entity.OrderItem{
		{
			ServiceID: svc.ID,
			Type:      entity.OrderItemTypeService,
			Quantity:  2,
			Variables: []entity.OrderItemVariable{
				{Name: "material", Type: "select", OptionValue: ptrStr("PLA")},
				{Name: "peso", Type: "number", NumberValue: ptrFloat(150)},
			},
		},
	}
	req := dto.UpdateOrderRequest{Items: &newItems}

	res, err := uc.UpdateOrder(context.Background(), "user-1", "order-2", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Items[0].UnitPrice != 300 {
		t.Errorf("expected recomputed unit_price=300, got %v", res.Items[0].UnitPrice)
	}
	if res.TotalAmount != 600 {
		t.Errorf("expected total=600 (300*2), got %v", res.TotalAmount)
	}
}
