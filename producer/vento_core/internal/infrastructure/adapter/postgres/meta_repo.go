package postgres

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type MetaRepo struct {
	db *sqlx.DB
}

func NewMetaRepo(db *sqlx.DB) *MetaRepo {
	return &MetaRepo{db: db}
}

func (r *MetaRepo) Save(ctx context.Context, config *entity.MetaConfig) error {
	query := `
		INSERT INTO meta_configs (user_id, platform_id, channel, whatsapp_business_id, permanent_access_token, verify_token, app_secret, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (platform_id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			channel = EXCLUDED.channel,
			whatsapp_business_id = EXCLUDED.whatsapp_business_id,
			permanent_access_token = EXCLUDED.permanent_access_token,
			verify_token = EXCLUDED.verify_token,
			app_secret = EXCLUDED.app_secret,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		config.UserID,
		config.PlatformID,
		config.Channel,
		config.WhatsAppBusinessID,
		config.PermanentAccessToken,
		config.VerifyToken,
		config.AppSecret,
	).Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)

	return err
}

func (r *MetaRepo) GetByUserID(ctx context.Context, userID string) (*entity.MetaConfig, error) {
	var config entity.MetaConfig
	query := `SELECT id, user_id, platform_id, channel, whatsapp_business_id, permanent_access_token, verify_token, app_secret, created_at, updated_at FROM meta_configs WHERE user_id = $1`
	err := r.db.GetContext(ctx, &config, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &config, err
}

func (r *MetaRepo) GetByPlatformID(ctx context.Context, platformID string) (*entity.MetaConfig, error) {
	var config entity.MetaConfig
	query := `SELECT id, user_id, platform_id, channel, whatsapp_business_id, permanent_access_token, verify_token, app_secret, created_at, updated_at FROM meta_configs WHERE platform_id = $1`
	err := r.db.GetContext(ctx, &config, query, platformID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &config, err
}
