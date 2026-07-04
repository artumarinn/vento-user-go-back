package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
)

type PostgresMetricsRepository struct {
	db *sqlx.DB
}

func NewPostgresMetricsRepository(db *sqlx.DB) *PostgresMetricsRepository {
	return &PostgresMetricsRepository{db: db}
}

func pctChange(current, previous float64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return ((current - previous) / previous) * 100
}

func (r *PostgresMetricsRepository) FetchMetrics(ctx context.Context, userID string, days int) (*dto.MetricsResponse, error) {
	now := time.Now().UTC()
	currentFrom := now.AddDate(0, 0, -days)
	previousFrom := now.AddDate(0, 0, -days*2)
	previousTo := currentFrom

	salesByDay, err := r.fetchSalesByDay(ctx, userID, currentFrom)
	if err != nil {
		return nil, err
	}

	topProducts, err := r.fetchTopProducts(ctx, userID, currentFrom)
	if err != nil {
		return nil, err
	}

	summary, err := r.fetchSummary(ctx, userID, currentFrom, previousFrom, previousTo)
	if err != nil {
		return nil, err
	}

	return &dto.MetricsResponse{
		SalesByDay:             salesByDay,
		TopProducts:            topProducts,
		ConversationsByChannel: []dto.ChannelCount{},
		ConversionFunnel:       []dto.FunnelStep{},
		Summary:                summary,
	}, nil
}

func (r *PostgresMetricsRepository) fetchSalesByDay(ctx context.Context, userID string, from time.Time) ([]dto.SalesByDayPoint, error) {
	type row struct {
		Name  string  `db:"name"`
		Value float64 `db:"value"`
	}
	var rows []row
	query := `
		SELECT
			TO_CHAR(paid_at AT TIME ZONE 'America/Argentina/Buenos_Aires', 'DD/MM') AS name,
			SUM(amount) AS value
		FROM order_payments
		WHERE user_id = $1 AND status = 'paid' AND paid_at >= $2
		GROUP BY name, DATE(paid_at AT TIME ZONE 'America/Argentina/Buenos_Aires')
		ORDER BY DATE(paid_at AT TIME ZONE 'America/Argentina/Buenos_Aires')
	`
	if err := r.db.SelectContext(ctx, &rows, query, userID, from); err != nil {
		return nil, err
	}
	result := make([]dto.SalesByDayPoint, len(rows))
	for i, rw := range rows {
		result[i] = dto.SalesByDayPoint{Name: rw.Name, Value: rw.Value}
	}
	return result, nil
}

func (r *PostgresMetricsRepository) fetchTopProducts(ctx context.Context, userID string, from time.Time) ([]dto.TopProduct, error) {
	type row struct {
		Name    string  `db:"name"`
		Units   int     `db:"units"`
		Revenue float64 `db:"revenue"`
	}
	var rows []row
	query := `
		SELECT
			item->>'name' AS name,
			SUM((item->>'quantity')::int) AS units,
			ROUND(SUM((item->>'quantity')::float * (item->>'unit_price')::float)::numeric, 0)::float AS revenue
		FROM orders, jsonb_array_elements(items) AS item
		WHERE user_id = $1 AND status != 'cancelled' AND created_at >= $2
		  AND item->>'name' IS NOT NULL AND item->>'name' != ''
		GROUP BY item->>'name'
		ORDER BY units DESC
		LIMIT 5
	`
	if err := r.db.SelectContext(ctx, &rows, query, userID, from); err != nil {
		return nil, err
	}
	result := make([]dto.TopProduct, len(rows))
	for i, rw := range rows {
		result[i] = dto.TopProduct{Name: rw.Name, Units: rw.Units, Revenue: rw.Revenue}
	}
	return result, nil
}

func (r *PostgresMetricsRepository) fetchSummary(ctx context.Context, userID string, currentFrom, previousFrom, previousTo time.Time) (dto.MetricsSummary, error) {
	var currentRevenue, previousRevenue, currentAvgTicket float64
	var currentOrders, previousOrders int

	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM order_payments WHERE user_id = $1 AND status = 'paid' AND paid_at >= $2`,
		userID, currentFrom,
	).Scan(&currentRevenue); err != nil {
		return dto.MetricsSummary{}, err
	}

	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM order_payments WHERE user_id = $1 AND status = 'paid' AND paid_at >= $2 AND paid_at < $3`,
		userID, previousFrom, previousTo,
	).Scan(&previousRevenue); err != nil {
		return dto.MetricsSummary{}, err
	}

	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM orders WHERE user_id = $1 AND status != 'cancelled' AND created_at >= $2`,
		userID, currentFrom,
	).Scan(&currentOrders); err != nil {
		return dto.MetricsSummary{}, err
	}

	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM orders WHERE user_id = $1 AND status != 'cancelled' AND created_at >= $2 AND created_at < $3`,
		userID, previousFrom, previousTo,
	).Scan(&previousOrders); err != nil {
		return dto.MetricsSummary{}, err
	}

	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(AVG(total), 0) FROM orders WHERE user_id = $1 AND status != 'cancelled' AND created_at >= $2`,
		userID, currentFrom,
	).Scan(&currentAvgTicket); err != nil {
		return dto.MetricsSummary{}, err
	}

	return dto.MetricsSummary{
		TotalRevenue:        currentRevenue,
		RevenueChange:       pctChange(currentRevenue, previousRevenue),
		TotalOrders:         currentOrders,
		OrdersChange:        pctChange(float64(currentOrders), float64(previousOrders)),
		TotalConversations:  0,
		ConversationsChange: 0,
		AvgTicket:           currentAvgTicket,
		TicketChange:        0,
	}, nil
}
