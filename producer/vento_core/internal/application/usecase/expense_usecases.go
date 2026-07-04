package usecase

import (
	"context"
	"time"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type ExpenseUsecases struct {
	repo port.ExpenseRepository
}

func NewExpenseUsecases(repo port.ExpenseRepository) *ExpenseUsecases {
	return &ExpenseUsecases{repo: repo}
}

func (uc *ExpenseUsecases) ListExpenses(ctx context.Context, userID string) ([]dto.ExpenseResponse, error) {
	expenses, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.ExpenseResponse, len(expenses))
	for i, e := range expenses {
		res[i] = mapExpenseEntityToDTO(e)
	}
	return res, nil
}

func (uc *ExpenseUsecases) CreateExpense(ctx context.Context, userID string, req dto.CreateExpenseRequest) (dto.ExpenseResponse, error) {
	paidAt := time.Now().UTC()
	if req.PaidAt != nil {
		paidAt = *req.PaidAt
	}

	e := entity.NewExpense(userID, req.Concept, req.Category, req.Amount, req.Method, paidAt)

	if err := uc.repo.Save(ctx, e); err != nil {
		return dto.ExpenseResponse{}, err
	}

	return mapExpenseEntityToDTO(e), nil
}

func (uc *ExpenseUsecases) DeleteExpense(ctx context.Context, userID string, expenseID string) error {
	return uc.repo.Delete(ctx, expenseID, userID)
}

func mapExpenseEntityToDTO(e *entity.Expense) dto.ExpenseResponse {
	return dto.ExpenseResponse{
		ID:        e.ID,
		Concept:   e.Concept,
		Category:  e.Category,
		Amount:    e.Amount,
		Method:    e.Method,
		PaidAt:    e.PaidAt,
		CreatedAt: e.CreatedAt,
	}
}
