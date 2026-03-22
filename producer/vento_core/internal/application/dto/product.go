package dto

import "time"

// CreateProductRequest is the input DTO for creating a new product.
type CreateProductRequest struct {
	Name         string   `json:"name" binding:"required"`
	SKU          string   `json:"sku" binding:"required"`
	Category     string   `json:"category"`
	Stock        float64  `json:"stock"`
	StockUnit    string   `json:"stock_unit"`
	MaxStock     *float64 `json:"max_stock"`
	Price        float64  `json:"price" binding:"required"`
	Supplier     string   `json:"supplier"`
	SupplierCost float64  `json:"supplier_cost"`
}

// UpdateProductRequest is the input DTO for updating an existing product.
type UpdateProductRequest struct {
	Name         string   `json:"name"`
	SKU          string   `json:"sku"`
	Category     string   `json:"category"`
	Stock        float64  `json:"stock"`
	StockUnit    string   `json:"stock_unit"`
	MaxStock     *float64 `json:"max_stock"`
	Price        float64  `json:"price"`
	Supplier     string   `json:"supplier"`
	SupplierCost float64  `json:"supplier_cost"`
}

// ProductResponse is the output DTO for product information.
type ProductResponse struct {
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
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
