package usecase

import (
	"context"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type PaymentRepository interface {
	ListByUserID(ctx context.Context, userID string) ([]*entity.Payment, error)
	Save(ctx context.Context, p *entity.Payment) error
}

type PaymentUsecases struct {
	repo PaymentRepository
}

func NewPaymentUsecases(repo PaymentRepository) *PaymentUsecases {
	return &PaymentUsecases{repo: repo}
}

func (uc *PaymentUsecases) ListPayments(ctx context.Context, userID string) ([]*entity.Payment, error) {
	return uc.repo.ListByUserID(ctx, userID)
}
