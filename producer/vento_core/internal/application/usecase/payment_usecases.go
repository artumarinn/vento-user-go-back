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
	repo port.PaymentRepository
}

func NewPaymentUsecases(repo port.PaymentRepository) *PaymentUsecases {
	return &PaymentUsecases{repo: repo}
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

	res := make([]dto.PaymentResponse, len(payments))
	for i, p := range payments {
		res[i] = dto.PaymentResponse{
			ID:         p.ID,
			UserID:     p.UserID,
			ClientName: p.ClientName,
			Amount:     p.Amount,
			Status:     string(p.Status),
			Concept:    p.Concept,
			CreatedAt:  p.CreatedAt,
			UpdatedAt:  p.UpdatedAt,
		}
	}
	return res, nil
}
