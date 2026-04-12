package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type OrderUsecases struct {
	repo port.OrderRepository
}

func NewOrderUsecases(repo port.OrderRepository) *OrderUsecases {
	return &OrderUsecases{repo: repo}
}

func (uc *OrderUsecases) SyncFromIA(ctx context.Context, userID string, req dto.BatchCreateOrderRequest) error {
	orders := make([]*entity.Order, len(req.Orders))
	for i, oReq := range req.Orders {
		orders[i] = &entity.Order{
			ID:             uuid.New().String(),
			UserID:         userID,
			ClientName:     oReq.ClientName,
			Total:          oReq.Total,
			Status:         entity.StatusPaymentReceived, // Default para IA sync
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
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
		res[i] = dto.OrderResponse{
			ID:             o.ID,
			UserID:         o.UserID,
			ClientName:     o.ClientName,
			ConversationID: o.ConversationID,
			Status:         string(o.Status),
			Total:          o.Total,
			CreatedAt:      o.CreatedAt,
			UpdatedAt:      o.UpdatedAt,
		}
	}
	return res, nil
}
