package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/vento-ai/shared/catalog"
)

// Service represents billable work priced via formula, owned by a user.
// Field shape is defined once in shared/catalog so it never drifts from
// vento-ai-service's copy.
type Service catalog.Service

// NewService creates a new service instance with a generated ID.
func NewService(userID, name, formula string, minimumLeadTime int) *Service {
	now := time.Now().UTC()
	return &Service{
		ID:              uuid.New().String(),
		UserID:          userID,
		Name:            name,
		Formula:         formula,
		MinimumLeadTime: minimumLeadTime,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
