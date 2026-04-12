package postgres

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type BusinessRepo struct {
	db *sqlx.DB
}

func NewBusinessRepo(db *sqlx.DB) *BusinessRepo {
	return &BusinessRepo{db: db}
}

func (r *BusinessRepo) Save(ctx context.Context, profile *entity.BusinessProfile) error {
	query := `
		INSERT INTO business_profiles (user_id, description, industry, tone, currency, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			description = EXCLUDED.description,
			industry = EXCLUDED.industry,
			tone = EXCLUDED.tone,
			currency = EXCLUDED.currency,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	// Primero verificamos si necesitamos el ON CONFLICT (user_id). 
	// Para eso, user_id debe ser UNIQUE en la tabla. 
	// Vamos a asegurar eso en el SQL después.
	
	err := r.db.QueryRowContext(ctx, query,
		profile.UserID,
		profile.Description,
		profile.Industry,
		profile.Tone,
		profile.Currency,
	).Scan(&profile.ID, &profile.CreatedAt, &profile.UpdatedAt)

	return err
}

func (r *BusinessRepo) GetByUserID(ctx context.Context, userID string) (*entity.BusinessProfile, error) {
	var profile entity.BusinessProfile
	query := `SELECT id, user_id, description, industry, tone, currency, created_at, updated_at FROM business_profiles WHERE user_id = $1`
	err := r.db.GetContext(ctx, &profile, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &profile, err
}

func (r *BusinessRepo) GetStats(ctx context.Context, userID string) (*entity.BusinessStats, error) {
	var stats entity.BusinessStats

	// Count Products
	err := r.db.GetContext(ctx, &stats.TotalProducts, "SELECT COUNT(*) FROM products WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	// Placeholder for Orders (since orders table might not exist yet)
	// In a real scenario, we would do:
	// SELECT SUM(total), COUNT(*) FILTER (WHERE status = 'seña_pagada') FROM orders WHERE user_id = $1
	stats.TotalSales = 0
	stats.PendingOrders = 0
	stats.ConversionRate = 0

	return &stats, nil
}
