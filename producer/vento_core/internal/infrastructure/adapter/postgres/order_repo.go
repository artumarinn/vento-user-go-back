package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type orderRow struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	ClientID       string    `db:"client_id"`
	ClientName     string    `db:"client_name"`
	ConversationID string    `db:"conversation_id"`
	Status         string    `db:"status"`
	Total          float64   `db:"total"`
	Items          []byte    `db:"items"` // JSONB in Postgres
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type PostgresOrderRepository struct {
	db *sqlx.DB
}

func NewPostgresOrderRepository(db *sqlx.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Order, error) {
	var rows []orderRow
	query := `SELECT * FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &rows, query, userID)
	if err != nil {
		return nil, err
	}

	orders := make([]*entity.Order, len(rows))
	for i, row := range rows {
		var items []entity.OrderItem
		_ = json.Unmarshal(row.Items, &items)

		orders[i] = &entity.Order{
			ID:             row.ID,
			UserID:         row.UserID,
			ClientID:       row.ClientID,
			ClientName:     row.ClientName,
			ConversationID: row.ConversationID,
			Status:         entity.OrderStatus(row.Status),
			Total:          row.Total,
			Items:          items,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		}
	}
	return orders, nil
}

func (r *PostgresOrderRepository) Save(ctx context.Context, o *entity.Order) error {
	itemsJSON, _ := json.Marshal(o.Items)
	query := `
		INSERT INTO orders (id, user_id, client_id, client_name, conversation_id, status, total, items, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		o.ID, o.UserID, o.ClientID, o.ClientName, o.ConversationID, string(o.Status), o.Total, itemsJSON, o.CreatedAt, o.UpdatedAt,
	)
	return err
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	var row orderRow
	query := `SELECT * FROM orders WHERE id = $1`
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		return nil, err
	}

	var items []entity.OrderItem
	_ = json.Unmarshal(row.Items, &items)

	return &entity.Order{
		ID:             row.ID,
		UserID:         row.UserID,
		ClientID:       row.ClientID,
		ClientName:     row.ClientName,
		ConversationID: row.ConversationID,
		Status:         entity.OrderStatus(row.Status),
		Total:          row.Total,
		Items:          items,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, userID string, status entity.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3 AND user_id = $4`
	_, err := r.db.ExecContext(ctx, query, string(status), time.Now(), id, userID)
	return err
}

func (r *PostgresOrderRepository) SaveBatch(ctx context.Context, orders []*entity.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO orders (id, user_id, client_id, client_name, conversation_id, status, total, items, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for _, o := range orders {
		itemsJSON, _ := json.Marshal(o.Items)
		_, err := tx.ExecContext(ctx, query,
			o.ID, o.UserID, o.ClientID, o.ClientName, o.ConversationID, string(o.Status), o.Total, itemsJSON, o.CreatedAt, o.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
