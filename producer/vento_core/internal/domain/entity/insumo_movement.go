package entity

import "time"

type InsumoMovement struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	InsumoID      string    `json:"insumo_id"`
	LocationID    string    `json:"location_id"`
	OrderID       *string   `json:"order_id"`
	QuantityDelta float64   `json:"quantity_delta"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
}
