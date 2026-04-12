package dto

import "time"

type ExtractedPaymentRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Contact  string `json:"contact"`
	Amount   float64 `json:"amount"` // Se agrega para gastos
}

type BatchCreatePaymentRequest struct {
	Vendors []ExtractedPaymentRequest `json:"vendors"`
	UserID  string                    `json:"user_id"`
}

type PaymentResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ClientName string    `json:"client_name"`
	Amount     float64   `json:"amount"`
	Status     string    `json:"status"`
	Concept    string    `json:"concept"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
