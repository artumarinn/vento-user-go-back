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

type OrderUsecases struct {
	repo        port.OrderRepository
	serviceRepo port.ServiceRepository
}

func NewOrderUsecases(repo port.OrderRepository, serviceRepo port.ServiceRepository) *OrderUsecases {
	return &OrderUsecases{repo: repo, serviceRepo: serviceRepo}
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

func (uc *OrderUsecases) ListOrders(ctx context.Context, userID string) ([]dto.OrderResponse, error) {
	orders, err := uc.repo.ListByUserID(ctx, userID)
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
	if err := uc.priceServiceItems(ctx, userID, req.Items); err != nil {
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

	if req.PartialAmount != nil && *req.PartialAmount > 0 {
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
	return uc.repo.Delete(ctx, orderID, userID)
}

func mapOrderEntityToResponse(o *entity.Order) dto.OrderResponse {
	var clientID *string
	if o.ClientID != "" {
		clientID = &o.ClientID
	}

	return dto.OrderResponse{
		ID:            o.ID,
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
