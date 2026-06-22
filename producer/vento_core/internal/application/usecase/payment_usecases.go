package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type PaymentUsecases struct {
	repo      port.PaymentRepository
	orderRepo port.OrderRepository
}

func NewPaymentUsecases(repo port.PaymentRepository, orderRepo port.OrderRepository) *PaymentUsecases {
	return &PaymentUsecases{repo: repo, orderRepo: orderRepo}
}

func orderPaymentConcept(kind entity.OrderPaymentKind, orderID string) string {
	switch kind {
	case entity.OrderPaymentKindDeposit:
		return "Seña pedido #" + orderID
	case entity.OrderPaymentKindBalance:
		return "Saldo pedido #" + orderID
	case entity.OrderPaymentKindPendingDebt:
		return "Fiado pendiente pedido #" + orderID
	default:
		return "Pago pedido #" + orderID
	}
}

func orderPaymentDisplayStatus(status entity.OrderPaymentStatus) string {
	if status == entity.OrderPaymentStatusPaid {
		return "completed"
	}
	return "pending"
}

func (uc *PaymentUsecases) SyncFromIA(ctx context.Context, userID string, req dto.BatchCreatePaymentRequest) error {
	payments := make([]*entity.Payment, len(req.Vendors))
	for i, vReq := range req.Vendors {
		payments[i] = &entity.Payment{
			ID:         uuid.New().String(),
			UserID:     userID,
			ClientName: vReq.Name,
			Amount:     vReq.Amount,
			Status:     "completed", // Pagos sync por IA (ej: gastos)
			Concept:    vReq.Category,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
	}
	return uc.repo.SaveBatch(ctx, payments)
}

func (uc *PaymentUsecases) ListPayments(ctx context.Context, userID string) ([]dto.PaymentResponse, error) {
	payments, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.PaymentResponse, 0, len(payments))
	for _, p := range payments {
		res = append(res, dto.PaymentResponse{
			ID:         p.ID,
			UserID:     p.UserID,
			ClientName: p.ClientName,
			Amount:     p.Amount,
			Status:     string(p.Status),
			Concept:    p.Concept,
			CreatedAt:  p.CreatedAt,
			UpdatedAt:  p.UpdatedAt,
		})
	}

	if uc.orderRepo == nil {
		return res, nil
	}

	orderPayments, err := uc.orderRepo.ListOrderPaymentsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, op := range orderPayments {
		orderID := op.OrderID
		res = append(res, dto.PaymentResponse{
			ID:         op.ID,
			UserID:     op.UserID,
			OrderID:    &orderID,
			ClientName: op.ClientName,
			Amount:     op.Amount,
			Method:     op.Method,
			Status:     orderPaymentDisplayStatus(op.Status),
			Concept:    orderPaymentConcept(op.Kind, op.OrderID),
			CreatedAt:  op.CreatedAt,
			UpdatedAt:  op.PaidAt,
		})
	}

	return res, nil
}
