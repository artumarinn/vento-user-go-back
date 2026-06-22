package dto

import "time"

// CreateInsumoRequest is the input DTO for creating a new insumo.
type CreateInsumoRequest struct {
	UserID    string   `json:"user_id"`
	Name      string   `json:"name" binding:"required"`
	Category  string   `json:"category"`
	Stock     *float64 `json:"stock"`
	StockUnit string   `json:"stock_unit"`
	Price     *float64 `json:"price"`
	Supplier  string   `json:"supplier"`
	TagIDs    []string `json:"tag_ids"`
}

// BatchCreateInsumoRequest is the input DTO for batch creation.
type BatchCreateInsumoRequest struct {
	Insumos []CreateInsumoRequest `json:"insumos"`
}

// UpdateInsumoRequest is the input DTO for updating an existing insumo.
type UpdateInsumoRequest struct {
	Name      string   `json:"name"`
	Category  string   `json:"category"`
	Stock     *float64 `json:"stock"`
	StockUnit string   `json:"stock_unit"`
	Price     *float64 `json:"price"`
	Supplier  string   `json:"supplier"`
	TagIDs    []string `json:"tag_ids"`
}

// InsumoResponse is the output DTO for insumo information.
type InsumoResponse struct {
	ID        string        `json:"id"`
	UserID    string        `json:"user_id"`
	Name      string        `json:"name"`
	Category  string        `json:"category"`
	Stock     *float64      `json:"stock"`
	StockUnit string        `json:"stock_unit"`
	Price     *float64      `json:"price"`
	Supplier  string        `json:"supplier"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Tags      []TagResponse `json:"tags"`
}
