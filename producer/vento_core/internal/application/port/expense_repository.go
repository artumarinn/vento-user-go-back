package port

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// ExpenseRepository defines the expected behavior for expense persistence.
type ExpenseRepository interface {
	Save(ctx context.Context, expense *entity.Expense) error
	Delete(ctx context.Context, id string, userID string) error
	ListByUserID(ctx context.Context, userID string) ([]*entity.Expense, error)
}
