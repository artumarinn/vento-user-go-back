package dto

import "time"

// CreateExpenseRequest is the input DTO for creating a new expense.
type CreateExpenseRequest struct {
	Concept  string     `json:"concept" binding:"required"`
	Category string     `json:"category"`
	Amount   float64    `json:"amount" binding:"required"`
	Method   string     `json:"method"`
	PaidAt   *time.Time `json:"paid_at"`
}

// ExpenseResponse is the output DTO for expense information.
type ExpenseResponse struct {
	ID        string    `json:"id"`
	Concept   string    `json:"concept"`
	Category  string    `json:"category"`
	Amount    float64   `json:"amount"`
	Method    string    `json:"method"`
	PaidAt    time.Time `json:"paid_at"`
	CreatedAt time.Time `json:"created_at"`
}
