package entity

import "time"

type OrderPaymentKind string

const (
	OrderPaymentKindDeposit     OrderPaymentKind = "deposit"
	OrderPaymentKindBalance     OrderPaymentKind = "balance"
	OrderPaymentKindFull        OrderPaymentKind = "full"
	OrderPaymentKindPendingDebt OrderPaymentKind = "pending_debt"
)

type OrderPaymentStatus string

const (
	OrderPaymentStatusPaid    OrderPaymentStatus = "paid"
	OrderPaymentStatusPending OrderPaymentStatus = "pending"
)

type OrderPayment struct {
	ID        string             `json:"id"`
	OrderID   string             `json:"order_id"`
	UserID    string             `json:"user_id"`
	Amount    float64            `json:"amount"`
	Method    string             `json:"method"`
	Kind      OrderPaymentKind   `json:"kind"`
	Status    OrderPaymentStatus `json:"status"`
	PaidAt    time.Time          `json:"paid_at"`
	CreatedAt time.Time          `json:"created_at"`
}

// OrderPaymentDetail enriches an OrderPayment with the client name of its order,
// used when listing payments across all orders for the Pagos panel.
type OrderPaymentDetail struct {
	OrderPayment
	ClientName string
}
