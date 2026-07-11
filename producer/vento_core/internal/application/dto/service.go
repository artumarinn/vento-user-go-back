package dto

import (
	"time"

	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// ServiceInsumoInput is the input DTO for declaring a recipe line when
// creating/updating a Service.
type ServiceInsumoInput struct {
	InsumoID        string  `json:"insumo_id" binding:"required"`
	QuantityPerUnit float64 `json:"quantity_per_unit" binding:"required"`
}

// ServiceInsumoResponse is the output DTO for a service's recipe line.
type ServiceInsumoResponse struct {
	InsumoID        string  `json:"insumo_id"`
	InsumoName      string  `json:"insumo_name"`
	QuantityPerUnit float64 `json:"quantity_per_unit"`
	Unit            string  `json:"unit"`
}

// CreateServiceRequest is the input DTO for creating a new service.
type CreateServiceRequest struct {
	UserID          string                       `json:"user_id"`
	Name            string                       `json:"name" binding:"required"`
	Formula         string                       `json:"formula" binding:"required"`
	MinimumLeadTime int                          `json:"minimum_lead_time"`
	TagIDs          []string                     `json:"tag_ids"`
	VariablesSchema []catalog.VariableDefinition `json:"variables_schema"`
	Insumos         []ServiceInsumoInput         `json:"insumos"`
}

// BatchCreateServiceRequest is the input DTO for batch creation.
type BatchCreateServiceRequest struct {
	Services []CreateServiceRequest `json:"services"`
}

// UpdateServiceRequest is the input DTO for updating an existing service.
type UpdateServiceRequest struct {
	Name            string                       `json:"name"`
	Formula         string                       `json:"formula"`
	MinimumLeadTime int                          `json:"minimum_lead_time"`
	TagIDs          []string                     `json:"tag_ids"`
	VariablesSchema []catalog.VariableDefinition `json:"variables_schema"`
	Insumos         []ServiceInsumoInput         `json:"insumos"`
}

// ServiceResponse is the output DTO for service information.
type ServiceResponse struct {
	ID              string                       `json:"id"`
	UserID          string                       `json:"user_id"`
	Name            string                       `json:"name"`
	Formula         string                       `json:"formula"`
	MinimumLeadTime int                          `json:"minimum_lead_time"`
	VariablesSchema []catalog.VariableDefinition `json:"variables_schema"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
	Tags            []TagResponse                `json:"tags"`
	Insumos         []ServiceInsumoResponse      `json:"insumos"`
}

// PreviewPriceRequest is the input DTO for the price-preview endpoint.
type PreviewPriceRequest struct {
	Variables []entity.OrderItemVariable `json:"variables"`
}

// PreviewPriceDraftRequest is the input DTO for previewing a price using an
// unsaved formula and variable schema (used while a Service is still being
// authored, before it has an ID).
type PreviewPriceDraftRequest struct {
	Formula         string                       `json:"formula" binding:"required"`
	VariablesSchema []catalog.VariableDefinition `json:"variables_schema"`
	Variables       []entity.OrderItemVariable   `json:"variables"`
}

// PreviewPriceResponse is the output DTO for the price-preview endpoint.
type PreviewPriceResponse struct {
	UnitPrice float64 `json:"unit_price"`
	LineTotal float64 `json:"line_total"`
}

// ServiceSearchResult is the customer-safe subset of a Service returned by
// the AI agent's search_services tool — never the formula or insumo recipe.
type ServiceSearchResult struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	MinimumLeadTime int    `json:"minimum_lead_time"`
}

// ServiceVariableOption is the customer-safe subset of a VariableOption —
// never unit_cost, which reveals internal pricing structure.
type ServiceVariableOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ServiceVariableResponse is the customer-safe subset of a
// catalog.VariableDefinition returned by the AI agent's get_service_variables
// tool — never unit_cost or anything formula-related.
type ServiceVariableResponse struct {
	Name     string                  `json:"name"`
	Label    string                  `json:"label"`
	Type     string                  `json:"type"`
	Unit     string                  `json:"unit,omitempty"`
	Options  []ServiceVariableOption `json:"options,omitempty"`
	Required bool                    `json:"required"`
}
