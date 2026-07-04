package entity

import "time"

// StockMovement is an audit entry for any change to a Product's Stock.
// QuantityDelta is positive when stock increases (cancellation, restock)
// and negative when stock decreases (sale). Reason is free text describing
// the cause: "sale", "cancellation", "adjustment".
type StockMovement struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ProductID     string    `json:"product_id"`
	LocationID    string    `json:"location_id"`
	OrderID       *string   `json:"order_id"`
	QuantityDelta float64   `json:"quantity_delta"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
}
