package entity

import (
	"time"

	"github.com/google/uuid"
)

// Location represents a physical or logical business location owned by a
// user. Each location has its own stock, orders and cash flow; catalog
// (products/services/prices) and clients stay shared across all locations.
type Location struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

// NewLocation creates a new location instance with a generated ID.
func NewLocation(userID, name, address, phone string, isDefault bool) *Location {
	return &Location{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		Address:   address,
		Phone:     phone,
		IsDefault: isDefault,
		CreatedAt: time.Now().UTC(),
	}
}
