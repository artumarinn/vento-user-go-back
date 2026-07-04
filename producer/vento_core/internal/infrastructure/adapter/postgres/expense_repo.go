package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type expenseRow struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Concept   string    `db:"concept"`
	Category  string    `db:"category"`
	Amount    float64   `db:"amount"`
	Method    string    `db:"method"`
	PaidAt    time.Time `db:"paid_at"`
	CreatedAt time.Time `db:"created_at"`
}

type PostgresExpenseRepository struct {
	db *sqlx.DB
}

func NewPostgresExpenseRepository(db *sqlx.DB) *PostgresExpenseRepository {
	return &PostgresExpenseRepository{db: db}
}

func (r *PostgresExpenseRepository) Save(ctx context.Context, e *entity.Expense) error {
	query := `
		INSERT INTO expenses (id, user_id, concept, category, amount, method, paid_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query, e.ID, e.UserID, e.Concept, e.Category, e.Amount, e.Method, e.PaidAt, e.CreatedAt)
	return err
}

func (r *PostgresExpenseRepository) Delete(ctx context.Context, id string, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM expenses WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *PostgresExpenseRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Expense, error) {
	var rows []expenseRow
	err := r.db.SelectContext(ctx, &rows, `SELECT * FROM expenses WHERE user_id = $1 ORDER BY paid_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	expenses := make([]*entity.Expense, len(rows))
	for i, row := range rows {
		expenses[i] = mapExpenseRowToEntity(row)
	}
	return expenses, nil
}

func mapExpenseRowToEntity(row expenseRow) *entity.Expense {
	return &entity.Expense{
		ID:        row.ID,
		UserID:    row.UserID,
		Concept:   row.Concept,
		Category:  row.Category,
		Amount:    row.Amount,
		Method:    row.Method,
		PaidAt:    row.PaidAt,
		CreatedAt: row.CreatedAt,
	}
}
