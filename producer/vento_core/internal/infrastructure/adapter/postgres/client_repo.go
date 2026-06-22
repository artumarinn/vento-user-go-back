package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type clientRow struct {
	ID               string         `db:"id"`
	UserID           string         `db:"user_id"`
	Name             string         `db:"name"`
	Phone            sql.NullString `db:"phone"`
	SocialNetwork    sql.NullString `db:"social_network"`
	SocialHandle     sql.NullString `db:"social_handle"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	TotalSpent       float64        `db:"total_spent"`
	OrderCount       int            `db:"order_count"`
	LastPurchaseDate sql.NullTime   `db:"last_purchase_date"`
}

type PostgresClientRepository struct {
	db *sqlx.DB
}

func NewPostgresClientRepository(db *sqlx.DB) *PostgresClientRepository {
	return &PostgresClientRepository{db: db}
}

func (r *PostgresClientRepository) Save(ctx context.Context, c *entity.Client) error {
	query := `
		INSERT INTO clients (id, user_id, name, phone, social_network, social_handle, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.UserID, c.Name, c.Phone, c.SocialNetwork, c.SocialHandle, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *PostgresClientRepository) ListByUserIDWithStats(ctx context.Context, userID string) ([]port.ClientStats, error) {
	query := `
		SELECT c.id, c.user_id, c.name, c.phone, c.social_network, c.social_handle, c.created_at, c.updated_at,
		       COALESCE(SUM(o.total), 0) AS total_spent,
		       COUNT(o.id) AS order_count,
		       MAX(o.created_at) AS last_purchase_date
		FROM clients c
		LEFT JOIN orders o ON o.client_id = c.id AND o.status != 'cancelled'
		WHERE c.user_id = $1
		GROUP BY c.id, c.user_id, c.name, c.phone, c.social_network, c.social_handle, c.created_at, c.updated_at
		ORDER BY c.created_at DESC
	`
	var rows []clientRow
	if err := r.db.SelectContext(ctx, &rows, query, userID); err != nil {
		return nil, err
	}

	stats := make([]port.ClientStats, len(rows))
	for i, row := range rows {
		phone := ""
		if row.Phone.Valid {
			phone = row.Phone.String
		}
		socialNetwork := ""
		if row.SocialNetwork.Valid {
			socialNetwork = row.SocialNetwork.String
		}
		socialHandle := ""
		if row.SocialHandle.Valid {
			socialHandle = row.SocialHandle.String
		}
		c := &entity.Client{
			ID:            row.ID,
			UserID:        row.UserID,
			Name:          row.Name,
			Phone:         phone,
			SocialNetwork: socialNetwork,
			SocialHandle:  socialHandle,
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		}

		var lastPurchase *time.Time
		if row.LastPurchaseDate.Valid {
			lastPurchase = &row.LastPurchaseDate.Time
		}

		stats[i] = port.ClientStats{
			Client:           c,
			TotalSpent:       row.TotalSpent,
			OrderCount:       row.OrderCount,
			LastPurchaseDate: lastPurchase,
		}
	}
	return stats, nil
}
