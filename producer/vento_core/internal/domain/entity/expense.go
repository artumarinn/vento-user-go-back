package entity

import (
	"time"

	"github.com/google/uuid"
)

// Expense represents money the business paid out (insumos, alquiler, luz,
// etc.) — the other half of "Pagos", which until now only tracked income.
type Expense struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Concept   string    `json:"concept"`
	Category  string    `json:"category"`
	Amount    float64   `json:"amount"`
	Method    string    `json:"method"`
	PaidAt    time.Time `json:"paid_at"`
	CreatedAt time.Time `json:"created_at"`
}

// NewExpense creates a new expense instance with a generated ID.
func NewExpense(userID, concept, category string, amount float64, method string, paidAt time.Time) *Expense {
	return &Expense{
		ID:        uuid.New().String(),
		UserID:    userID,
		Concept:   concept,
		Category:  category,
		Amount:    amount,
		Method:    method,
		PaidAt:    paidAt,
		CreatedAt: time.Now().UTC(),
	}
}
