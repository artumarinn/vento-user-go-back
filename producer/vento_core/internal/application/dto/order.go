package dto

import (
	"time"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type ExtractedOrderRequest struct {
	ClientName   string   `json:"client_name"`
	Total        *float64 `json:"total"`
	Status       string   `json:"status"`
	Date         string   `json:"date"`
	ItemsSummary string   `json:"items_summary"`
}

type BatchCreateOrderRequest struct {
	Orders []ExtractedOrderRequest `json:"orders"`
	UserID string                  `json:"user_id"`
}

type CreateOrderRequest struct {
	ClientID      *string            `json:"client_id"`
	ClientName    string             `json:"client_name"`
	Channel       string             `json:"channel"`
	Items         []entity.OrderItem `json:"items"`
	DeliveryDate  *string            `json:"delivery_date"`
	PaymentMethod string             `json:"payment_method"`
	PaymentStatus string             `json:"payment_status"`
	PartialAmount *float64           `json:"partial_amount"`
	Notes         string             `json:"notes"`
	Status        string             `json:"status"`
}

type UpdateOrderRequest struct {
	ClientID      *string             `json:"client_id"`
	ClientName    *string             `json:"client_name"`
	Channel       *string             `json:"channel"`
	Items         *[]entity.OrderItem `json:"items"`
	DeliveryDate  *string             `json:"delivery_date"`
	PaymentMethod *string             `json:"payment_method"`
	PaymentStatus *string             `json:"payment_status"`
	PartialAmount *float64            `json:"partial_amount"`
	Notes         *string             `json:"notes"`
	Status        *string             `json:"status"`
}

type OrderResponse struct {
	ID            string             `json:"id"`
	ClientID      *string            `json:"client_id"`
	ClientName    string             `json:"client_name"`
	Channel       string             `json:"channel"`
	Items         []entity.OrderItem `json:"items"`
	ItemsCount    int                `json:"items_count"`
	TotalAmount   float64            `json:"total_amount"`
	DeliveryDate  *string            `json:"delivery_date"`
	PaymentMethod string             `json:"payment_method"`
	PaymentStatus string             `json:"payment_status"`
	Notes         string             `json:"notes"`
	Status        string             `json:"status"`
	Date          time.Time          `json:"date"`
}

type OrderStatusEventResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type PaymentInfoResponse struct {
	Method string    `json:"method"`
	Amount float64   `json:"amount"`
	Kind   string    `json:"kind"`
	Status string    `json:"status"`
	PaidAt time.Time `json:"paid_at"`
}

type OrderDetailResponse struct {
	OrderResponse
	StatusHistory     []OrderStatusEventResponse `json:"status_history"`
	ConversationID    *string                    `json:"conversation_id"`
	Payments          []PaymentInfoResponse      `json:"payments"`
	AmountPaid        float64                    `json:"amount_paid"`
	AmountOutstanding float64                    `json:"amount_outstanding"`
}

type RegisterPaymentRequest struct {
	Method string  `json:"method"`
	Amount float64 `json:"amount"`
}
