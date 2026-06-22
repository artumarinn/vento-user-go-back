package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type tagRow struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Label     string    `db:"label"`
	Color     string    `db:"color"`
	CreatedAt time.Time `db:"created_at"`
}

type PostgresTagRepository struct {
	db *sqlx.DB
}

func NewPostgresTagRepository(db *sqlx.DB) *PostgresTagRepository {
	return &PostgresTagRepository{db: db}
}

func (r *PostgresTagRepository) Save(ctx context.Context, t *entity.Tag) error {
	query := `
		INSERT INTO tags (id, user_id, label, color, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, label) DO UPDATE SET
			color = EXCLUDED.color
	`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.UserID, t.Label, t.Color, t.CreatedAt)
	return err
}

func (r *PostgresTagRepository) Delete(ctx context.Context, id string, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tags WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *PostgresTagRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Tag, error) {
	var rows []tagRow
	err := r.db.SelectContext(ctx, &rows, `SELECT * FROM tags WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	tags := make([]*entity.Tag, len(rows))
	for i, row := range rows {
		tags[i] = mapTagRowToEntity(row)
	}
	return tags, nil
}

func (r *PostgresTagRepository) ListByInsumoID(ctx context.Context, insumoID string) ([]*entity.Tag, error) {
	var rows []tagRow
	query := `
		SELECT t.* FROM tags t
		JOIN insumo_tags it ON it.tag_id = t.id
		WHERE it.insumo_id = $1
	`
	err := r.db.SelectContext(ctx, &rows, query, insumoID)
	if err != nil {
		return nil, err
	}
	tags := make([]*entity.Tag, len(rows))
	for i, row := range rows {
		tags[i] = mapTagRowToEntity(row)
	}
	return tags, nil
}

func (r *PostgresTagRepository) SetInsumoTags(ctx context.Context, insumoID string, tagIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM insumo_tags WHERE insumo_id = $1`, insumoID); err != nil {
		return err
	}

	if len(tagIDs) > 0 {
		query := `INSERT INTO insumo_tags (insumo_id, tag_id) VALUES ($1, $2)`
		for _, tagID := range tagIDs {
			if _, err := tx.ExecContext(ctx, query, insumoID, tagID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *PostgresTagRepository) ListByProductID(ctx context.Context, productID string) ([]*entity.Tag, error) {
	var rows []tagRow
	query := `
		SELECT t.* FROM tags t
		JOIN product_tags pt ON pt.tag_id = t.id
		WHERE pt.product_id = $1
	`
	err := r.db.SelectContext(ctx, &rows, query, productID)
	if err != nil {
		return nil, err
	}
	tags := make([]*entity.Tag, len(rows))
	for i, row := range rows {
		tags[i] = mapTagRowToEntity(row)
	}
	return tags, nil
}

func (r *PostgresTagRepository) SetProductTags(ctx context.Context, productID string, tagIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM product_tags WHERE product_id = $1`, productID); err != nil {
		return err
	}

	if len(tagIDs) > 0 {
		query := `INSERT INTO product_tags (product_id, tag_id) VALUES ($1, $2)`
		for _, tagID := range tagIDs {
			if _, err := tx.ExecContext(ctx, query, productID, tagID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *PostgresTagRepository) ListByServiceID(ctx context.Context, serviceID string) ([]*entity.Tag, error) {
	var rows []tagRow
	query := `
		SELECT t.* FROM tags t
		JOIN service_tags st ON st.tag_id = t.id
		WHERE st.service_id = $1
	`
	err := r.db.SelectContext(ctx, &rows, query, serviceID)
	if err != nil {
		return nil, err
	}
	tags := make([]*entity.Tag, len(rows))
	for i, row := range rows {
		tags[i] = mapTagRowToEntity(row)
	}
	return tags, nil
}

func (r *PostgresTagRepository) SetServiceTags(ctx context.Context, serviceID string, tagIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM service_tags WHERE service_id = $1`, serviceID); err != nil {
		return err
	}

	if len(tagIDs) > 0 {
		query := `INSERT INTO service_tags (service_id, tag_id) VALUES ($1, $2)`
		for _, tagID := range tagIDs {
			if _, err := tx.ExecContext(ctx, query, serviceID, tagID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func mapTagRowToEntity(row tagRow) *entity.Tag {
	return &entity.Tag{
		ID:        row.ID,
		UserID:    row.UserID,
		Label:     row.Label,
		Color:     row.Color,
		CreatedAt: row.CreatedAt,
	}
}
