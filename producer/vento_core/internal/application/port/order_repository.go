package port

import (
	"context"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type OrderRepository interface {
	Save(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id string) (*entity.Order, error)
	GetByIDForUser(ctx context.Context, id string, userID string) (*entity.Order, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.Order, error)
	UpdateStatus(ctx context.Context, id string, userID string, status entity.OrderStatus) error
	SaveBatch(ctx context.Context, orders []*entity.Order) error
	Update(ctx context.Context, order *entity.Order) error
	Delete(ctx context.Context, id, userID string) error
	AppendStatusHistory(ctx context.Context, orderID string, status entity.OrderStatus) error
	GetStatusHistory(ctx context.Context, orderID string) ([]entity.OrderStatusEvent, error)
	RegisterPayment(ctx context.Context, orderID string, method string, amount float64) error
	InsertOrderPayment(ctx context.Context, payment *entity.OrderPayment) error
	ListOrderPayments(ctx context.Context, orderID string) ([]entity.OrderPayment, error)
	ListOrderPaymentsByUserID(ctx context.Context, userID string) ([]entity.OrderPaymentDetail, error)
	SumPaidOrderPayments(ctx context.Context, orderID string) (float64, error)
	UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus string) error
}
