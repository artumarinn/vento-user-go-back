package entity

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a catalog item owned by a user.
type Product struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	SKU          string    `json:"sku"`
	Category     string    `json:"category"`
	Stock        float64   `json:"stock"`
	StockUnit    string    `json:"stock_unit"`
	MaxStock     *float64  `json:"max_stock,omitempty"`
	Price        float64   `json:"price"`
	Supplier     string    `json:"supplier"`
	SupplierCost float64   `json:"supplier_cost"`
	Description  string    `json:"description"`
	Tags         string    `json:"tags"`
	ImageURL     string    `json:"image_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	TagRefs      []Tag     `json:"tag_refs"`
}

// NewProduct creates a new product instance with a generated ID.
func NewProduct(userID, name, sku, category string, price float64) *Product {
	now := time.Now().UTC()
	return &Product{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		SKU:       sku,
		Category:  category,
		Price:     price,
		StockUnit: "u",
		CreatedAt: now,
		UpdatedAt: now,
	}
}
