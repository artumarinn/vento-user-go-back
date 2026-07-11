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
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if profile.BusinessName != "" {
		_, err = tx.ExecContext(ctx, `UPDATE users SET business_name = $1 WHERE id = $2`, profile.BusinessName, profile.UserID)
		if err != nil {
			return err
		}
	}

	// default_agent_mode and tone are intentionally NOT written here — they
	// are owned by UpdateAgentMode/UpdateTone now. Writing them
	// unconditionally used to blank each field every time "Mi Negocio"
	// saved the rest of the profile, since that form no longer sends them.
	query := `
		INSERT INTO business_profiles (user_id, business_name, description, industry, currency, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			business_name = EXCLUDED.business_name,
			description   = EXCLUDED.description,
			industry      = EXCLUDED.industry,
			currency      = EXCLUDED.currency,
			updated_at    = NOW()
		RETURNING id, default_agent_mode, tone, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, query,
		profile.UserID,
		profile.BusinessName,
		profile.Description,
		profile.Industry,
		profile.Currency,
	).Scan(&profile.ID, &profile.DefaultAgentMode, &profile.Tone, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateAgentMode sets ONLY the default_agent_mode column, never touching the
// other profile fields. The "Cómo responde" toggle uses this instead of the
// full Save upsert — Save overwrites every unspecified column with its zero
// value, which was blanking the whole business profile every time the mode
// was flipped (the data-loss bug). Upserts so it also works before a profile
// has ever been saved.
func (r *BusinessRepo) UpdateAgentMode(ctx context.Context, userID, mode string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO business_profiles (user_id, default_agent_mode, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			default_agent_mode = EXCLUDED.default_agent_mode,
			updated_at         = NOW()
	`, userID, mode)
	return err
}

// UpdateTone sets ONLY the tone column, never touching the other profile
// fields — same isolation as UpdateAgentMode, and for the same reason: "Mi
// Negocio" no longer sends tone in its payload, so Save's full upsert must
// never be the one that changes it.
func (r *BusinessRepo) UpdateTone(ctx context.Context, userID, tone string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO business_profiles (user_id, tone, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			tone       = EXCLUDED.tone,
			updated_at = NOW()
	`, userID, tone)
	return err
}

func (r *BusinessRepo) GetByUserID(ctx context.Context, userID string) (*entity.BusinessProfile, error) {
	var profile entity.BusinessProfile
	query := `SELECT id, user_id, business_name, description, industry, tone, currency, default_agent_mode, created_at, updated_at FROM business_profiles WHERE user_id = $1`
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
