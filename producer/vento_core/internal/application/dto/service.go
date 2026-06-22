package dto

import (
	"time"

	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// CreateServiceRequest is the input DTO for creating a new service.
type CreateServiceRequest struct {
	UserID          string                       `json:"user_id"`
	Name            string                       `json:"name" binding:"required"`
	Formula         string                       `json:"formula" binding:"required"`
	MinimumLeadTime int                          `json:"minimum_lead_time"`
	TagIDs          []string                     `json:"tag_ids"`
	VariablesSchema []catalog.VariableDefinition `json:"variables_schema"`
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
}

// PreviewPriceRequest is the input DTO for the price-preview endpoint.
type PreviewPriceRequest struct {
	Variables []entity.OrderItemVariable `json:"variables"`
}

// PreviewPriceResponse is the output DTO for the price-preview endpoint.
type PreviewPriceResponse struct {
	UnitPrice float64 `json:"unit_price"`
	LineTotal float64 `json:"line_total"`
}
