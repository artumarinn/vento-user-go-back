package entity

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusProcessing OrderStatus = "processing"
	StatusReady      OrderStatus = "ready"
	StatusDelivered  OrderStatus = "delivered"
	StatusPaused     OrderStatus = "paused"
	StatusCancelled  OrderStatus = "cancelled"
)

type OrderItemType string

const (
	OrderItemTypeProduct OrderItemType = "product"
	OrderItemTypeService OrderItemType = "service"
)

type OrderItemVariable struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	OptionValue  *string  `json:"option_value,omitempty"`
	NumberValue  *float64 `json:"number_value,omitempty"`
	TextValue    *string  `json:"text_value,omitempty"`
	BooleanValue *bool    `json:"boolean_value,omitempty"`
	PricedValue  *float64 `json:"priced_value,omitempty"`
}

type OrderItem struct {
	ProductID string              `json:"product_id"`
	ServiceID string              `json:"service_id,omitempty"`
	Type      OrderItemType       `json:"type,omitempty"`
	Name      string              `json:"name"`
	Quantity  int                 `json:"quantity"`
	UnitPrice float64             `json:"unit_price"`
	Variables []OrderItemVariable `json:"variables,omitempty"`
}

type Order struct {
	ID                    string      `json:"id"`
	UserID                string      `json:"user_id"`
	LocationID            string      `json:"location_id"`
	ClientID              string      `json:"client_id"`
	ClientName            string      `json:"client_name"`
	ConversationID        string      `json:"conversation_id"`
	Status                OrderStatus `json:"status"`
	Total                 float64     `json:"total"`
	Items                 []OrderItem `json:"items"`
	Channel               string      `json:"channel"`
	DeliveryDate          *string     `json:"delivery_date"`
	PaymentMethod         string      `json:"payment_method"`
	PaymentStatus         string      `json:"payment_status"`
	PartialAmount         *float64    `json:"partial_amount"`
	Notes                 string      `json:"notes"`
	PaymentRecordedMethod *string     `json:"payment_recorded_method"`
	PaymentRecordedAmount *float64    `json:"payment_recorded_amount"`
	PaymentRecordedAt     *time.Time  `json:"payment_recorded_at"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
}

type OrderStatusEvent struct {
	Status    OrderStatus
	ChangedAt time.Time
}

func NewOrder(userID, clientName, conversationID string, items []OrderItem) *Order {
	var total float64
	for _, item := range items {
		total += item.UnitPrice * float64(item.Quantity)
	}

	return &Order{
		ID:             uuid.New().String(),
		UserID:         userID,
		ClientName:     clientName,
		ConversationID: conversationID,
		Status:         StatusPending,
		Total:          total,
		Items:          items,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
