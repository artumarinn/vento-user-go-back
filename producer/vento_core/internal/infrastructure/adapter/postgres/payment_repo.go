package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type paymentRow struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	ClientName     string    `db:"client_name"`
	Amount         float64   `db:"amount"`
	Status         string    `db:"status"`
	Channel        string    `db:"channel"`
	Concept        string    `db:"concept"`
	ComprobanteURL string    `db:"comprobante_url"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type PostgresPaymentRepository struct {
	db *sqlx.DB
}

func NewPostgresPaymentRepository(db *sqlx.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Payment, error) {
	var rows []paymentRow
	query := `SELECT * FROM payments WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &rows, query, userID)
	if err != nil {
		return nil, err
	}

	payments := make([]*entity.Payment, len(rows))
	for i, row := range rows {
		payments[i] = &entity.Payment{
			ID:             row.ID,
			UserID:         row.UserID,
			ClientName:     row.ClientName,
			Amount:         row.Amount,
			Status:         entity.PaymentStatus(row.Status),
			Channel:        row.Channel,
			Concept:        row.Concept,
			ComprobanteURL: row.ComprobanteURL,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		}
	}
	return payments, nil
}

func (r *PostgresPaymentRepository) Save(ctx context.Context, p *entity.Payment) error {
	query := `
		INSERT INTO payments (id, user_id, client_name, amount, status, channel, concept, comprobante_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.UserID, p.ClientName, p.Amount, string(p.Status), p.Channel, p.Concept, p.ComprobanteURL, p.CreatedAt, p.UpdatedAt,
	)
	return err
}
