package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/pricing"
)

var validOrderStatuses = map[string]bool{
	string(entity.StatusPending):    true,
	string(entity.StatusConfirmed):  true,
	string(entity.StatusProcessing): true,
	string(entity.StatusReady):      true,
	string(entity.StatusDelivered):  true,
	string(entity.StatusPaused):     true,
	string(entity.StatusCancelled):  true,
}

// ErrOrderLockedByPayment indicates an order cannot have its Service item
// variables/items mutated because at least one OrderPayment already exists.
// Persisted prices are frozen once a payment is registered (golden rule:
// price is computed once by code, never recalculated retroactively).
var ErrOrderLockedByPayment = errors.New("order items locked: payment already registered")

// ErrInsufficientStock indicates an order item requests more units of a
// product than are currently available. Stock is never allowed to go
// negative — the sale is rejected instead.
var ErrInsufficientStock = errors.New("order: insufficient stock")

// ErrInvalidOrderItems is returned when an order has no items or any item
// has a quantity of zero or less.
var ErrInvalidOrderItems = errors.New("order: must have at least one item with quantity > 0")

// ErrLocationNotOwnedByUser is returned when a request supplies a
// location_id that does not belong to the authenticated user. Prevents one
// tenant from mutating another tenant's location_stock/movements ledger by
// guessing/reusing a foreign location UUID.
var ErrLocationNotOwnedByUser = errors.New("order: location does not belong to this user")

func validateOrderItems(items []entity.OrderItem) error {
	if len(items) == 0 {
		return ErrInvalidOrderItems
	}
	for _, item := range items {
		if item.Quantity <= 0 {
			return ErrInvalidOrderItems
		}
	}
	return nil
}

type OrderUsecases struct {
	repo         port.OrderRepository
	serviceRepo  port.ServiceRepository
	productRepo  port.ProductRepository
	insumoRepo   port.InsumoRepository
	locationRepo port.LocationRepository
}

func NewOrderUsecases(repo port.OrderRepository, serviceRepo port.ServiceRepository, productRepo port.ProductRepository, insumoRepo port.InsumoRepository, locationRepo port.LocationRepository) *OrderUsecases {
	return &OrderUsecases{repo: repo, serviceRepo: serviceRepo, productRepo: productRepo, insumoRepo: insumoRepo, locationRepo: locationRepo}
}

// resolveOwnedLocationID validates that a client-supplied location_id
// belongs to userID, returning ErrLocationNotOwnedByUser otherwise. An empty
// rawLocationID (order without a resolved location) is passed through
// unchanged — callers such as SyncFromIA/internal flows may not have one.
func (uc *OrderUsecases) resolveOwnedLocationID(ctx context.Context, userID string, rawLocationID *string) (string, error) {
	if rawLocationID == nil || *rawLocationID == "" {
		return "", nil
	}
	loc, err := uc.locationRepo.GetByID(ctx, *rawLocationID, userID)
	if err != nil {
		return "", err
	}
	if loc == nil {
		return "", ErrLocationNotOwnedByUser
	}
	return loc.ID, nil
}

// isProductItem reports whether an order item represents a Product (as
// opposed to a Service). Legacy orders never set Type, so the absence of
// OrderItemTypeService combined with a non-empty ProductID is treated as a
// product item.
func isProductItem(item entity.OrderItem) bool {
	return item.ProductID != "" && item.Type != entity.OrderItemTypeService
}

// priceServiceItems computes UnitPrice for each Service-type item via the
// pricing evaluator, overwriting any client-supplied unit_price. Non-service
// items are left untouched.
func (uc *OrderUsecases) priceServiceItems(ctx context.Context, userID string, items []entity.OrderItem) error {
	for i := range items {
		item := &items[i]
		if item.Type != entity.OrderItemTypeService {
			continue
		}

		service, err := uc.serviceRepo.GetByID(ctx, item.ServiceID, userID)
		if err != nil {
			return err
		}
		if service == nil {
			return fmt.Errorf("service not found: %s", item.ServiceID)
		}

		vars, err := buildVarMap(service, item.Variables)
		if err != nil {
			return err
		}

		price, err := pricing.Evaluate(service.Formula, vars)
		if err != nil {
			return err
		}
		item.UnitPrice = price
	}
	return nil
}

// validateProductStock checks every product item against currently available
// stock and fails fast (rejecting the whole sale) if any single item would
// drive a product's stock below zero. No stock is touched here.
func (uc *OrderUsecases) validateProductStock(ctx context.Context, userID string, items []entity.OrderItem) error {
	for _, item := range items {
		if !isProductItem(item) {
			continue
		}
		product, err := uc.productRepo.GetByID(ctx, item.ProductID, userID)
		if err != nil {
			return err
		}
		if product == nil {
			return fmt.Errorf("product not found: %s", item.ProductID)
		}
		if product.Stock < float64(item.Quantity) {
			return fmt.Errorf("%w: %s", ErrInsufficientStock, product.Name)
		}
	}
	return nil
}

// decrementProductStock applies a "sale" stock movement for every product
// item in the order. Must only be called after the order itself has been
// persisted successfully and after validateProductStock has passed.
// locationID mirrors the movement into location_stock; empty when the order
// has no resolved location (legacy/internal callers).
func (uc *OrderUsecases) decrementProductStock(ctx context.Context, userID, orderID, locationID string, items []entity.OrderItem, now time.Time) error {
	for _, item := range items {
		if !isProductItem(item) {
			continue
		}
		if err := uc.productRepo.AdjustStock(ctx, item.ProductID, userID, -float64(item.Quantity)); err != nil {
			return err
		}
		if err := uc.productRepo.AdjustLocationStock(ctx, locationID, item.ProductID, -float64(item.Quantity)); err != nil {
			return err
		}
		movement := &entity.StockMovement{
			UserID:        userID,
			ProductID:     item.ProductID,
			LocationID:    locationID,
			OrderID:       &orderID,
			QuantityDelta: -float64(item.Quantity),
			Reason:        "sale",
			CreatedAt:     now,
		}
		if err := uc.productRepo.InsertStockMovement(ctx, movement); err != nil {
			return err
		}
	}
	return nil
}

// restoreProductStock applies a stock movement that gives back units of
// every product item in the order (cancellation/deletion path).
func (uc *OrderUsecases) restoreProductStock(ctx context.Context, userID, orderID, locationID string, items []entity.OrderItem, reason string, now time.Time) error {
	for _, item := range items {
		if !isProductItem(item) {
			continue
		}
		if err := uc.productRepo.AdjustStock(ctx, item.ProductID, userID, float64(item.Quantity)); err != nil {
			return err
		}
		if err := uc.productRepo.AdjustLocationStock(ctx, locationID, item.ProductID, float64(item.Quantity)); err != nil {
			return err
		}
		movement := &entity.StockMovement{
			UserID:        userID,
			ProductID:     item.ProductID,
			LocationID:    locationID,
			OrderID:       &orderID,
			QuantityDelta: float64(item.Quantity),
			Reason:        reason,
			CreatedAt:     now,
		}
		if err := uc.productRepo.InsertStockMovement(ctx, movement); err != nil {
			return err
		}
	}
	return nil
}

// productQuantities sums Quantity per ProductID across product-type items,
// ignoring service items entirely.
func productQuantities(items []entity.OrderItem) map[string]float64 {
	qty := make(map[string]float64)
	for _, item := range items {
		if !isProductItem(item) {
			continue
		}
		qty[item.ProductID] += float64(item.Quantity)
	}
	return qty
}

// applyProductStockDelta reconciles stock for an order edit: for each
// product whose requested quantity changed between oldItems and newItems,
// it validates (when the new quantity is higher) and applies an "adjustment"
// stock movement for just the difference. Validation for every increased
// product happens before any stock is touched, so a rejected edit never
// partially adjusts stock.
func (uc *OrderUsecases) applyProductStockDelta(ctx context.Context, userID, orderID, locationID string, oldItems, newItems []entity.OrderItem, now time.Time) error {
	oldQty := productQuantities(oldItems)
	newQty := productQuantities(newItems)

	productIDs := make(map[string]struct{}, len(oldQty)+len(newQty))
	for id := range oldQty {
		productIDs[id] = struct{}{}
	}
	for id := range newQty {
		productIDs[id] = struct{}{}
	}

	deltas := make(map[string]float64, len(productIDs))
	for id := range productIDs {
		delta := newQty[id] - oldQty[id]
		if delta == 0 {
			continue
		}
		if delta > 0 {
			product, err := uc.productRepo.GetByID(ctx, id, userID)
			if err != nil {
				return err
			}
			if product == nil {
				return fmt.Errorf("product not found: %s", id)
			}
			if product.Stock < delta {
				return fmt.Errorf("%w: %s", ErrInsufficientStock, product.Name)
			}
		}
		deltas[id] = delta
	}

	for id, delta := range deltas {
		if err := uc.productRepo.AdjustStock(ctx, id, userID, -delta); err != nil {
			return err
		}
		if err := uc.productRepo.AdjustLocationStock(ctx, locationID, id, -delta); err != nil {
			return err
		}
		movement := &entity.StockMovement{
			UserID:        userID,
			ProductID:     id,
			LocationID:    locationID,
			OrderID:       &orderID,
			QuantityDelta: -delta,
			Reason:        "adjustment",
			CreatedAt:     now,
		}
		if err := uc.productRepo.InsertStockMovement(ctx, movement); err != nil {
			return err
		}
	}
	return nil
}

// serviceInsumoConsumption sums, per insumo_id, how many units a set of
// order items would consume according to each Service's recipe. Only
// service-type items are considered — products never consume insumos
// directly.
func (uc *OrderUsecases) serviceInsumoConsumption(ctx context.Context, items []entity.OrderItem) (map[string]float64, error) {
	consumption := make(map[string]float64)
	for _, item := range items {
		if item.Type != entity.OrderItemTypeService {
			continue
		}
		recipe, err := uc.serviceRepo.ListServiceInsumos(ctx, item.ServiceID)
		if err != nil {
			return nil, err
		}
		for _, line := range recipe {
			consumption[line.InsumoID] += line.QuantityPerUnit * float64(item.Quantity)
		}
	}
	return consumption, nil
}

// validateInsumoStock checks every insumo consumed by the order's service
// items against currently available stock and fails fast (rejecting the
// whole sale) if any insumo would go negative. No stock is touched here.
// An insumo with Stock == nil was never loaded with stock data and must
// never block a sale.
func (uc *OrderUsecases) validateInsumoStock(ctx context.Context, userID string, items []entity.OrderItem) error {
	consumption, err := uc.serviceInsumoConsumption(ctx, items)
	if err != nil {
		return err
	}
	for insumoID, qty := range consumption {
		insumo, err := uc.insumoRepo.GetByID(ctx, insumoID, userID)
		if err != nil {
			return err
		}
		if insumo == nil {
			continue
		}
		if insumo.Stock != nil && *insumo.Stock < qty {
			return fmt.Errorf("%w: insumo %s", ErrInsufficientStock, insumo.Name)
		}
	}
	return nil
}

// decrementInsumoStock applies a "production" stock movement for every
// insumo consumed by the order's service items. Must only be called after
// the order itself has been persisted successfully and after
// validateInsumoStock has passed.
func (uc *OrderUsecases) decrementInsumoStock(ctx context.Context, userID, orderID, locationID string, items []entity.OrderItem, now time.Time) error {
	consumption, err := uc.serviceInsumoConsumption(ctx, items)
	if err != nil {
		return err
	}
	for insumoID, qty := range consumption {
		if qty == 0 {
			continue
		}
		if err := uc.insumoRepo.AdjustStock(ctx, insumoID, userID, -qty); err != nil {
			return err
		}
		if err := uc.insumoRepo.AdjustLocationInsumoStock(ctx, locationID, insumoID, -qty); err != nil {
			return err
		}
		movement := &entity.InsumoMovement{
			UserID: userID, InsumoID: insumoID, LocationID: locationID, OrderID: &orderID,
			QuantityDelta: -qty, Reason: "production", CreatedAt: now,
		}
		if err := uc.insumoRepo.InsertInsumoMovement(ctx, movement); err != nil {
			return err
		}
	}
	return nil
}

// restoreInsumoStock applies a stock movement that gives back units of
// every insumo consumed by the order's service items (cancellation/deletion
// path).
func (uc *OrderUsecases) restoreInsumoStock(ctx context.Context, userID, orderID, locationID string, items []entity.OrderItem, reason string, now time.Time) error {
	consumption, err := uc.serviceInsumoConsumption(ctx, items)
	if err != nil {
		return err
	}
	for insumoID, qty := range consumption {
		if qty == 0 {
			continue
		}
		if err := uc.insumoRepo.AdjustStock(ctx, insumoID, userID, qty); err != nil {
			return err
		}
		if err := uc.insumoRepo.AdjustLocationInsumoStock(ctx, locationID, insumoID, qty); err != nil {
			return err
		}
		movement := &entity.InsumoMovement{
			UserID: userID, InsumoID: insumoID, LocationID: locationID, OrderID: &orderID,
			QuantityDelta: qty, Reason: reason, CreatedAt: now,
		}
		if err := uc.insumoRepo.InsertInsumoMovement(ctx, movement); err != nil {
			return err
		}
	}
	return nil
}

// applyInsumoStockDelta reconciles insumo stock for an order edit: for each
// insumo whose total consumption changed between oldItems and newItems, it
// validates (when consumption increases) and applies an "adjustment" stock
// movement for just the difference. Validation for every increased insumo
// happens before any stock is touched, so a rejected edit never partially
// adjusts stock.
func (uc *OrderUsecases) applyInsumoStockDelta(ctx context.Context, userID, orderID, locationID string, oldItems, newItems []entity.OrderItem, now time.Time) error {
	oldConsumption, err := uc.serviceInsumoConsumption(ctx, oldItems)
	if err != nil {
		return err
	}
	newConsumption, err := uc.serviceInsumoConsumption(ctx, newItems)
	if err != nil {
		return err
	}

	insumoIDs := make(map[string]struct{}, len(oldConsumption)+len(newConsumption))
	for id := range oldConsumption {
		insumoIDs[id] = struct{}{}
	}
	for id := range newConsumption {
		insumoIDs[id] = struct{}{}
	}

	deltas := make(map[string]float64, len(insumoIDs))
	for id := range insumoIDs {
		delta := newConsumption[id] - oldConsumption[id]
		if delta == 0 {
			continue
		}
		if delta > 0 {
			insumo, err := uc.insumoRepo.GetByID(ctx, id, userID)
			if err != nil {
				return err
			}
			if insumo != nil && insumo.Stock != nil && *insumo.Stock < delta {
				return fmt.Errorf("%w: insumo %s", ErrInsufficientStock, insumo.Name)
			}
		}
		deltas[id] = delta
	}

	for id, delta := range deltas {
		if err := uc.insumoRepo.AdjustStock(ctx, id, userID, -delta); err != nil {
			return err
		}
		if err := uc.insumoRepo.AdjustLocationInsumoStock(ctx, locationID, id, -delta); err != nil {
			return err
		}
		movement := &entity.InsumoMovement{
			UserID: userID, InsumoID: id, LocationID: locationID, OrderID: &orderID,
			QuantityDelta: -delta, Reason: "adjustment", CreatedAt: now,
		}
		if err := uc.insumoRepo.InsertInsumoMovement(ctx, movement); err != nil {
			return err
		}
	}
	return nil
}

func (uc *OrderUsecases) SyncFromIA(ctx context.Context, userID string, req dto.BatchCreateOrderRequest) error {
	orders := make([]*entity.Order, len(req.Orders))
	for i, oReq := range req.Orders {
		total := 0.0
		if oReq.Total != nil {
			total = *oReq.Total
		}

		orders[i] = &entity.Order{
			ID:         uuid.New().String(),
			UserID:     userID,
			ClientName: oReq.ClientName,
			Total:      total,
			Status:     entity.StatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
	}
	return uc.repo.SaveBatch(ctx, orders)
}

// ListOrders returns orders for a user. When locationID is empty, orders
// across every location are returned (aggregate "Todo el negocio" view).
func (uc *OrderUsecases) ListOrders(ctx context.Context, userID string, locationID string) ([]dto.OrderResponse, error) {
	var orders []*entity.Order
	var err error
	if locationID != "" {
		orders, err = uc.repo.ListByUserIDAndLocation(ctx, userID, locationID)
	} else {
		orders, err = uc.repo.ListByUserID(ctx, userID)
	}
	if err != nil {
		return nil, err
	}

	res := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		res[i] = mapOrderEntityToResponse(o)
	}
	return res, nil
}

func (uc *OrderUsecases) CreateOrder(ctx context.Context, userID string, req dto.CreateOrderRequest) (dto.OrderResponse, error) {
	if err := validateOrderItems(req.Items); err != nil {
		return dto.OrderResponse{}, err
	}

	// A client-supplied location_id must belong to this tenant — validated
	// up front, before any pricing/stock work, so a spoofed foreign location
	// never touches another tenant's ledger.
	locationID, err := uc.resolveOwnedLocationID(ctx, userID, req.LocationID)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	if err := uc.priceServiceItems(ctx, userID, req.Items); err != nil {
		return dto.OrderResponse{}, err
	}

	// Validate stock availability for every product item BEFORE creating the
	// order — a sale is rejected in full if any item exceeds stock, never
	// partially applied.
	if err := uc.validateProductStock(ctx, userID, req.Items); err != nil {
		return dto.OrderResponse{}, err
	}

	if err := uc.validateInsumoStock(ctx, userID, req.Items); err != nil {
		return dto.OrderResponse{}, err
	}

	var total float64
	for _, item := range req.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}

	clientID := ""
	if req.ClientID != nil {
		clientID = *req.ClientID
	}

	now := time.Now()
	o := &entity.Order{
		ID:            uuid.New().String(),
		UserID:        userID,
		LocationID:    locationID,
		ClientID:      clientID,
		ClientName:    req.ClientName,
		Status:        entity.OrderStatus(req.Status),
		Total:         total,
		Items:         req.Items,
		Channel:       req.Channel,
		DeliveryDate:  req.DeliveryDate,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: req.PaymentStatus,
		PartialAmount: req.PartialAmount,
		Notes:         req.Notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := uc.repo.Save(ctx, o); err != nil {
		return dto.OrderResponse{}, err
	}

	if err := uc.decrementProductStock(ctx, userID, o.ID, o.LocationID, o.Items, now); err != nil {
		return dto.OrderResponse{}, err
	}

	if err := uc.decrementInsumoStock(ctx, userID, o.ID, o.LocationID, o.Items, now); err != nil {
		return dto.OrderResponse{}, err
	}

	// A "paid"/"partial" payment_status at creation time must leave behind an
	// actual OrderPayment record, otherwise amount_paid (derived by summing
	// payments) silently diverges from the status label.
	switch {
	case req.PaymentStatus == "paid":
		payment := &entity.OrderPayment{
			OrderID:   o.ID,
			UserID:    userID,
			Amount:    total,
			Method:    req.PaymentMethod,
			Kind:      entity.OrderPaymentKindFull,
			Status:    entity.OrderPaymentStatusPaid,
			PaidAt:    now,
			CreatedAt: now,
		}
		if err := uc.repo.InsertOrderPayment(ctx, payment); err != nil {
			return dto.OrderResponse{}, err
		}
	case req.PartialAmount != nil && *req.PartialAmount > 0:
		payment := &entity.OrderPayment{
			OrderID:   o.ID,
			UserID:    userID,
			Amount:    *req.PartialAmount,
			Method:    req.PaymentMethod,
			Kind:      entity.OrderPaymentKindDeposit,
			Status:    entity.OrderPaymentStatusPaid,
			PaidAt:    now,
			CreatedAt: now,
		}
		if err := uc.repo.InsertOrderPayment(ctx, payment); err != nil {
			return dto.OrderResponse{}, err
		}
	}

	return mapOrderEntityToResponse(o), nil
}

func (uc *OrderUsecases) GetOrderDetail(ctx context.Context, userID, orderID string) (*dto.OrderDetailResponse, error) {
	o, err := uc.repo.GetByIDForUser(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, nil
	}

	history, err := uc.repo.GetStatusHistory(ctx, orderID)
	if err != nil {
		return nil, err
	}

	historyRes := make([]dto.OrderStatusEventResponse, len(history))
	for i, h := range history {
		historyRes[i] = dto.OrderStatusEventResponse{Status: string(h.Status), Timestamp: h.ChangedAt}
	}

	var conversationID *string
	if o.ConversationID != "" {
		conversationID = &o.ConversationID
	}

	orderPayments, err := uc.repo.ListOrderPayments(ctx, orderID)
	if err != nil {
		return nil, err
	}

	paymentsRes := make([]dto.PaymentInfoResponse, len(orderPayments))
	var amountPaid float64
	for i, p := range orderPayments {
		paymentsRes[i] = dto.PaymentInfoResponse{
			Method: p.Method,
			Amount: p.Amount,
			Kind:   string(p.Kind),
			Status: string(p.Status),
			PaidAt: p.PaidAt,
		}
		if p.Status == entity.OrderPaymentStatusPaid {
			amountPaid += p.Amount
		}
	}
	amountOutstanding := o.Total - amountPaid
	if amountOutstanding < 0 {
		amountOutstanding = 0
	}

	return &dto.OrderDetailResponse{
		OrderResponse:     mapOrderEntityToResponse(o),
		StatusHistory:     historyRes,
		ConversationID:    conversationID,
		Payments:          paymentsRes,
		AmountPaid:        amountPaid,
		AmountOutstanding: amountOutstanding,
	}, nil
}

func (uc *OrderUsecases) UpdateOrderStatus(ctx context.Context, userID, orderID string, status string) (dto.OrderResponse, error) {
	if !validOrderStatuses[status] {
		return dto.OrderResponse{}, fmt.Errorf("invalid status: %s", status)
	}

	current, err := uc.repo.GetByIDForUser(ctx, orderID, userID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	if current == nil {
		return dto.OrderResponse{}, fmt.Errorf("order not found: %s", orderID)
	}
	if current.Status == entity.StatusDelivered || current.Status == entity.StatusCancelled {
		return dto.OrderResponse{}, fmt.Errorf("cannot change status of a %s order", current.Status)
	}

	if status == string(entity.StatusDelivered) {
		if current.PaymentStatus != "paid" && current.PaymentStatus != "credit" {
			return dto.OrderResponse{}, fmt.Errorf("cannot complete order: outstanding balance, register payment first")
		}
		if current.PaymentStatus == "credit" {
			totalPaid, err := uc.repo.SumPaidOrderPayments(ctx, orderID)
			if err != nil {
				return dto.OrderResponse{}, err
			}
			remaining := current.Total - totalPaid
			if remaining > 0 {
				now := time.Now()
				debt := &entity.OrderPayment{
					OrderID:   orderID,
					UserID:    userID,
					Amount:    remaining,
					Method:    current.PaymentMethod,
					Kind:      entity.OrderPaymentKindPendingDebt,
					Status:    entity.OrderPaymentStatusPending,
					PaidAt:    now,
					CreatedAt: now,
				}
				if err := uc.repo.InsertOrderPayment(ctx, debt); err != nil {
					return dto.OrderResponse{}, err
				}
			}
		}
	}

	if err := uc.repo.UpdateStatus(ctx, orderID, userID, entity.OrderStatus(status)); err != nil {
		return dto.OrderResponse{}, err
	}

	if status == string(entity.StatusCancelled) {
		if err := uc.restoreProductStock(ctx, userID, orderID, current.LocationID, current.Items, "cancellation", time.Now().UTC()); err != nil {
			return dto.OrderResponse{}, err
		}
		if err := uc.restoreInsumoStock(ctx, userID, orderID, current.LocationID, current.Items, "cancellation", time.Now().UTC()); err != nil {
			return dto.OrderResponse{}, err
		}
	}

	o, err := uc.repo.GetByIDForUser(ctx, orderID, userID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	if o == nil {
		return dto.OrderResponse{}, fmt.Errorf("order not found: %s", orderID)
	}
	return mapOrderEntityToResponse(o), nil
}

// computeOrderPaymentStatus derives the new payment_status after a payment is registered.
// "credit" only persists while no payment has been made yet — once any amount is paid,
// the order behaves like any other (partial/paid) regardless of how it started.
func computeOrderPaymentStatus(currentStatus string, totalPaid, total float64) string {
	if total > 0 && totalPaid >= total {
		return "paid"
	}
	if totalPaid > 0 {
		return "partial"
	}
	if currentStatus == "credit" {
		return "credit"
	}
	return "pending"
}

func (uc *OrderUsecases) RegisterPayment(ctx context.Context, userID, orderID string, req dto.RegisterPaymentRequest) (dto.OrderResponse, error) {
	o, err := uc.repo.GetByIDForUser(ctx, orderID, userID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	if o == nil {
		return dto.OrderResponse{}, fmt.Errorf("order not found: %s", orderID)
	}

	existing, err := uc.repo.ListOrderPayments(ctx, orderID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	kind := entity.OrderPaymentKindFull
	if len(existing) > 0 {
		kind = entity.OrderPaymentKindBalance
	}

	now := time.Now()
	payment := &entity.OrderPayment{
		OrderID:   orderID,
		UserID:    userID,
		Amount:    req.Amount,
		Method:    req.Method,
		Kind:      kind,
		Status:    entity.OrderPaymentStatusPaid,
		PaidAt:    now,
		CreatedAt: now,
	}
	if err := uc.repo.InsertOrderPayment(ctx, payment); err != nil {
		return dto.OrderResponse{}, err
	}

	totalPaid, err := uc.repo.SumPaidOrderPayments(ctx, orderID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	newStatus := computeOrderPaymentStatus(o.PaymentStatus, totalPaid, o.Total)
	if err := uc.repo.UpdatePaymentStatus(ctx, orderID, newStatus); err != nil {
		return dto.OrderResponse{}, err
	}

	updated, err := uc.repo.GetByIDForUser(ctx, orderID, userID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	if updated == nil {
		return dto.OrderResponse{}, fmt.Errorf("order not found: %s", orderID)
	}
	return mapOrderEntityToResponse(updated), nil
}

func (uc *OrderUsecases) UpdateOrder(ctx context.Context, userID, orderID string, req dto.UpdateOrderRequest) (dto.OrderResponse, error) {
	o, err := uc.repo.GetByIDForUser(ctx, orderID, userID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	if o == nil {
		return dto.OrderResponse{}, fmt.Errorf("order not found: %s", orderID)
	}

	if req.ClientID != nil {
		o.ClientID = *req.ClientID
	}
	if req.ClientName != nil {
		o.ClientName = *req.ClientName
	}
	if req.Channel != nil {
		o.Channel = *req.Channel
	}
	if req.Items != nil {
		if err := validateOrderItems(*req.Items); err != nil {
			return dto.OrderResponse{}, err
		}

		existingPayments, err := uc.repo.ListOrderPayments(ctx, orderID)
		if err != nil {
			return dto.OrderResponse{}, err
		}
		if len(existingPayments) > 0 {
			return dto.OrderResponse{}, ErrOrderLockedByPayment
		}

		newItems := *req.Items
		if err := uc.priceServiceItems(ctx, userID, newItems); err != nil {
			return dto.OrderResponse{}, err
		}

		if err := uc.applyProductStockDelta(ctx, userID, orderID, o.LocationID, o.Items, newItems, time.Now().UTC()); err != nil {
			return dto.OrderResponse{}, err
		}

		if err := uc.applyInsumoStockDelta(ctx, userID, orderID, o.LocationID, o.Items, newItems, time.Now().UTC()); err != nil {
			return dto.OrderResponse{}, err
		}

		o.Items = newItems
		var total float64
		for _, item := range o.Items {
			total += item.UnitPrice * float64(item.Quantity)
		}
		o.Total = total
	}
	if req.DeliveryDate != nil {
		o.DeliveryDate = req.DeliveryDate
	}
	if req.PaymentMethod != nil {
		o.PaymentMethod = *req.PaymentMethod
	}
	if req.PaymentStatus != nil {
		o.PaymentStatus = *req.PaymentStatus
	}
	if req.PartialAmount != nil {
		o.PartialAmount = req.PartialAmount
	}
	if req.Notes != nil {
		o.Notes = *req.Notes
	}
	if req.Status != nil {
		o.Status = entity.OrderStatus(*req.Status)
	}
	o.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, o); err != nil {
		return dto.OrderResponse{}, err
	}

	return mapOrderEntityToResponse(o), nil
}

func (uc *OrderUsecases) DeleteOrder(ctx context.Context, userID, orderID string) error {
	existing, err := uc.repo.GetByIDForUser(ctx, orderID, userID)
	if err != nil {
		return err
	}
	// Stock is only restored if the order was never cancelled — a cancelled
	// order already got its stock back via UpdateOrderStatus, so restoring
	// it again here would double-credit the product.
	if existing != nil && existing.Status != entity.StatusCancelled {
		if err := uc.restoreProductStock(ctx, userID, orderID, existing.LocationID, existing.Items, "cancellation", time.Now().UTC()); err != nil {
			return err
		}
		if err := uc.restoreInsumoStock(ctx, userID, orderID, existing.LocationID, existing.Items, "cancellation", time.Now().UTC()); err != nil {
			return err
		}
	}
	return uc.repo.Delete(ctx, orderID, userID)
}

func mapOrderEntityToResponse(o *entity.Order) dto.OrderResponse {
	var clientID *string
	if o.ClientID != "" {
		clientID = &o.ClientID
	}

	return dto.OrderResponse{
		ID:            o.ID,
		LocationID:    o.LocationID,
		ClientID:      clientID,
		ClientName:    o.ClientName,
		Channel:       o.Channel,
		Items:         o.Items,
		ItemsCount:    len(o.Items),
		TotalAmount:   o.Total,
		DeliveryDate:  o.DeliveryDate,
		PaymentMethod: o.PaymentMethod,
		PaymentStatus: o.PaymentStatus,
		Notes:         o.Notes,
		Status:        string(o.Status),
		Date:          o.CreatedAt,
	}
}
