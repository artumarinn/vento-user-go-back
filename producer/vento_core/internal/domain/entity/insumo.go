package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/vento-ai/shared/catalog"
)

// Insumo represents a raw material item owned by a user.
// Field shape is defined once in shared/catalog so it never drifts from
// vento-ai-service's copy.
type Insumo catalog.Insumo

// NewInsumo creates a new insumo instance with a generated ID.
func NewInsumo(userID, name, category string, price float64) *Insumo {
	now := time.Now().UTC()
	return &Insumo{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		Category:  category,
		Price:     &price,
		StockUnit: "u",
		CreatedAt: now,
		UpdatedAt: now,
	}
}
