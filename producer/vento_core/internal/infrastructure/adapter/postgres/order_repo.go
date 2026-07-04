package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type orderRow struct {
	ID                    string          `db:"id"`
	UserID                string          `db:"user_id"`
	LocationID            sql.NullString  `db:"location_id"`
	ClientID              string          `db:"client_id"`
	ClientName            string          `db:"client_name"`
	ConversationID        string          `db:"conversation_id"`
	Status                string          `db:"status"`
	Total                 float64         `db:"total"`
	Items                 []byte          `db:"items"` // JSONB in Postgres
	Channel               string          `db:"channel"`
	DeliveryDate          sql.NullString  `db:"delivery_date"`
	PaymentMethod         string          `db:"payment_method"`
	PaymentStatus         string          `db:"payment_status"`
	PartialAmount         sql.NullFloat64 `db:"partial_amount"`
	Notes                 sql.NullString  `db:"notes"`
	PaymentRecordedMethod sql.NullString  `db:"payment_recorded_method"`
	PaymentRecordedAmount sql.NullFloat64 `db:"payment_recorded_amount"`
	PaymentRecordedAt     sql.NullTime    `db:"payment_recorded_at"`
	CreatedAt             time.Time       `db:"created_at"`
	UpdatedAt             time.Time       `db:"updated_at"`
}

func (row orderRow) toEntity() *entity.Order {
	var items []entity.OrderItem
	_ = json.Unmarshal(row.Items, &items)

	o := &entity.Order{
		ID:             row.ID,
		UserID:         row.UserID,
		LocationID:     row.LocationID.String,
		ClientID:       row.ClientID,
		ClientName:     row.ClientName,
		ConversationID: row.ConversationID,
		Status:         entity.OrderStatus(row.Status),
		Total:          row.Total,
		Items:          items,
		Channel:        row.Channel,
		PaymentMethod:  row.PaymentMethod,
		PaymentStatus:  row.PaymentStatus,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
	if row.DeliveryDate.Valid {
		o.DeliveryDate = &row.DeliveryDate.String
	}
	if row.PartialAmount.Valid {
		o.PartialAmount = &row.PartialAmount.Float64
	}
	if row.Notes.Valid {
		o.Notes = row.Notes.String
	}
	if row.PaymentRecordedMethod.Valid {
		o.PaymentRecordedMethod = &row.PaymentRecordedMethod.String
	}
	if row.PaymentRecordedAmount.Valid {
		o.PaymentRecordedAmount = &row.PaymentRecordedAmount.Float64
	}
	if row.PaymentRecordedAt.Valid {
		o.PaymentRecordedAt = &row.PaymentRecordedAt.Time
	}
	return o
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
		orders[i] = row.toEntity()
	}
	return orders, nil
}

func (r *PostgresOrderRepository) ListByUserIDAndLocation(ctx context.Context, userID string, locationID string) ([]*entity.Order, error) {
	var rows []orderRow
	query := `SELECT * FROM orders WHERE user_id = $1 AND location_id = $2 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &rows, query, userID, locationID)
	if err != nil {
		return nil, err
	}

	orders := make([]*entity.Order, len(rows))
	for i, row := range rows {
		orders[i] = row.toEntity()
	}
	return orders, nil
}

func (r *PostgresOrderRepository) Save(ctx context.Context, o *entity.Order) error {
	itemsJSON, _ := json.Marshal(o.Items)
	query := `
		INSERT INTO orders (id, user_id, location_id, client_id, client_name, conversation_id, status, total, items, channel, delivery_date, payment_method, payment_status, partial_amount, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`
	_, err := r.db.ExecContext(ctx, query,
		o.ID, o.UserID, nullableUUID(o.LocationID), o.ClientID, o.ClientName, o.ConversationID, string(o.Status), o.Total, itemsJSON,
		o.Channel, o.DeliveryDate, o.PaymentMethod, o.PaymentStatus, o.PartialAmount, o.Notes, o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return r.AppendStatusHistory(ctx, o.ID, o.Status)
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	var row orderRow
	query := `SELECT * FROM orders WHERE id = $1`
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		return nil, err
	}
	return row.toEntity(), nil
}

func (r *PostgresOrderRepository) GetByIDForUser(ctx context.Context, id string, userID string) (*entity.Order, error) {
	var row orderRow
	query := `SELECT * FROM orders WHERE id = $1 AND user_id = $2`
	err := r.db.GetContext(ctx, &row, query, id, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return row.toEntity(), nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, userID string, status entity.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3 AND user_id = $4`
	_, err := r.db.ExecContext(ctx, query, string(status), time.Now(), id, userID)
	if err != nil {
		return err
	}
	return r.AppendStatusHistory(ctx, id, status)
}

func (r *PostgresOrderRepository) Update(ctx context.Context, o *entity.Order) error {
	itemsJSON, _ := json.Marshal(o.Items)
	query := `
		UPDATE orders SET
			client_id = $1, client_name = $2, channel = $3, items = $4, total = $5,
			delivery_date = $6, payment_method = $7, payment_status = $8, partial_amount = $9,
			notes = $10, status = $11, updated_at = $12
		WHERE id = $13 AND user_id = $14
	`
	_, err := r.db.ExecContext(ctx, query,
		o.ClientID, o.ClientName, o.Channel, itemsJSON, o.Total,
		o.DeliveryDate, o.PaymentMethod, o.PaymentStatus, o.PartialAmount,
		o.Notes, string(o.Status), o.UpdatedAt, o.ID, o.UserID,
	)
	return err
}

func (r *PostgresOrderRepository) Delete(ctx context.Context, id, userID string) error {
	query := `DELETE FROM orders WHERE id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}

func (r *PostgresOrderRepository) AppendStatusHistory(ctx context.Context, orderID string, status entity.OrderStatus) error {
	query := `INSERT INTO order_status_history (order_id, status, changed_at) VALUES ($1, $2, NOW())`
	_, err := r.db.ExecContext(ctx, query, orderID, string(status))
	return err
}

func (r *PostgresOrderRepository) GetStatusHistory(ctx context.Context, orderID string) ([]entity.OrderStatusEvent, error) {
	type historyRow struct {
		Status    string    `db:"status"`
		ChangedAt time.Time `db:"changed_at"`
	}
	var rows []historyRow
	query := `SELECT status, changed_at FROM order_status_history WHERE order_id = $1 ORDER BY changed_at ASC`
	err := r.db.SelectContext(ctx, &rows, query, orderID)
	if err != nil {
		return nil, err
	}

	events := make([]entity.OrderStatusEvent, len(rows))
	for i, row := range rows {
		events[i] = entity.OrderStatusEvent{Status: entity.OrderStatus(row.Status), ChangedAt: row.ChangedAt}
	}
	return events, nil
}

func (r *PostgresOrderRepository) RegisterPayment(ctx context.Context, orderID string, method string, amount float64) error {
	query := `UPDATE orders SET payment_recorded_method = $1, payment_recorded_amount = $2, payment_recorded_at = NOW(), updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, method, amount, orderID)
	return err
}

func (r *PostgresOrderRepository) InsertOrderPayment(ctx context.Context, payment *entity.OrderPayment) error {
	query := `
		INSERT INTO order_payments (order_id, user_id, amount, method, kind, status, paid_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.OrderID, payment.UserID, payment.Amount, payment.Method,
		string(payment.Kind), string(payment.Status), payment.PaidAt, payment.CreatedAt,
	)
	return err
}

type orderPaymentRow struct {
	ID        string    `db:"id"`
	OrderID   string    `db:"order_id"`
	UserID    string    `db:"user_id"`
	Amount    float64   `db:"amount"`
	Method    string    `db:"method"`
	Kind      string    `db:"kind"`
	Status    string    `db:"status"`
	PaidAt    time.Time `db:"paid_at"`
	CreatedAt time.Time `db:"created_at"`
}

func (r *PostgresOrderRepository) ListOrderPayments(ctx context.Context, orderID string) ([]entity.OrderPayment, error) {
	var rows []orderPaymentRow
	query := `SELECT id, order_id, user_id, amount, method, kind, status, paid_at, created_at FROM order_payments WHERE order_id = $1 ORDER BY paid_at ASC`
	if err := r.db.SelectContext(ctx, &rows, query, orderID); err != nil {
		return nil, err
	}
	payments := make([]entity.OrderPayment, len(rows))
	for i, row := range rows {
		payments[i] = entity.OrderPayment{
			ID:        row.ID,
			OrderID:   row.OrderID,
			UserID:    row.UserID,
			Amount:    row.Amount,
			Method:    row.Method,
			Kind:      entity.OrderPaymentKind(row.Kind),
			Status:    entity.OrderPaymentStatus(row.Status),
			PaidAt:    row.PaidAt,
			CreatedAt: row.CreatedAt,
		}
	}
	return payments, nil
}

type orderPaymentWithClientRow struct {
	ID         string    `db:"id"`
	OrderID    string    `db:"order_id"`
	UserID     string    `db:"user_id"`
	Amount     float64   `db:"amount"`
	Method     string    `db:"method"`
	Kind       string    `db:"kind"`
	Status     string    `db:"status"`
	PaidAt     time.Time `db:"paid_at"`
	CreatedAt  time.Time `db:"created_at"`
	ClientName string    `db:"client_name"`
}

func (r *PostgresOrderRepository) ListOrderPaymentsByUserID(ctx context.Context, userID string) ([]entity.OrderPaymentDetail, error) {
	var rows []orderPaymentWithClientRow
	query := `
		SELECT op.id, op.order_id, op.user_id, op.amount, op.method, op.kind, op.status, op.paid_at, op.created_at, o.client_name
		FROM order_payments op
		JOIN orders o ON o.id = op.order_id
		WHERE op.user_id = $1
		ORDER BY op.paid_at DESC
	`
	if err := r.db.SelectContext(ctx, &rows, query, userID); err != nil {
		return nil, err
	}
	details := make([]entity.OrderPaymentDetail, len(rows))
	for i, row := range rows {
		details[i] = entity.OrderPaymentDetail{
			OrderPayment: entity.OrderPayment{
				ID:        row.ID,
				OrderID:   row.OrderID,
				UserID:    row.UserID,
				Amount:    row.Amount,
				Method:    row.Method,
				Kind:      entity.OrderPaymentKind(row.Kind),
				Status:    entity.OrderPaymentStatus(row.Status),
				PaidAt:    row.PaidAt,
				CreatedAt: row.CreatedAt,
			},
			ClientName: row.ClientName,
		}
	}
	return details, nil
}

func (r *PostgresOrderRepository) SumPaidOrderPayments(ctx context.Context, orderID string) (float64, error) {
	var sum sql.NullFloat64
	query := `SELECT SUM(amount) FROM order_payments WHERE order_id = $1 AND status = 'paid'`
	if err := r.db.GetContext(ctx, &sum, query, orderID); err != nil {
		return 0, err
	}
	return sum.Float64, nil
}

func (r *PostgresOrderRepository) UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus string) error {
	query := `UPDATE orders SET payment_status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, paymentStatus, orderID)
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
		if _, err := tx.ExecContext(ctx, `INSERT INTO order_status_history (order_id, status, changed_at) VALUES ($1, $2, $3)`, o.ID, string(o.Status), o.CreatedAt); err != nil {
			return err
		}
	}

	return tx.Commit()
}
