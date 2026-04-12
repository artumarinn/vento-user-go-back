package dto

import "time"

type ExtractedOrderRequest struct {
	ClientName   string  `json:"client_name"`
	Total        float64 `json:"total"`
	Status       string  `json:"status"`
	Date         string  `json:"date"`
	ItemsSummary string  `json:"items_summary"`
}

type BatchCreateOrderRequest struct {
	Orders []ExtractedOrderRequest `json:"orders"`
	UserID string                  `json:"user_id"`
}

type OrderResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	ClientName     string    `json:"client_name"`
	ConversationID string    `json:"conversation_id"`
	Status         string    `json:"status"`
	Total          float64   `json:"total"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
