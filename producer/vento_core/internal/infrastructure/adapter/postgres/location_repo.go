package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type locationRow struct {
	ID        string         `db:"id"`
	UserID    string         `db:"user_id"`
	Name      string         `db:"name"`
	Address   sql.NullString `db:"address"`
	Phone     sql.NullString `db:"phone"`
	IsDefault bool           `db:"is_default"`
	CreatedAt time.Time      `db:"created_at"`
}

type PostgresLocationRepository struct {
	db *sqlx.DB
}

func NewPostgresLocationRepository(db *sqlx.DB) *PostgresLocationRepository {
	return &PostgresLocationRepository{db: db}
}

func (r *PostgresLocationRepository) Save(ctx context.Context, l *entity.Location) error {
	query := `
		INSERT INTO locations (id, user_id, name, address, phone, is_default, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		l.ID, l.UserID, l.Name, l.Address, l.Phone, l.IsDefault, l.CreatedAt,
	)
	return err
}

func (r *PostgresLocationRepository) Update(ctx context.Context, l *entity.Location) error {
	query := `
		UPDATE locations
		SET name = $1, address = $2, phone = $3, is_default = $4
		WHERE id = $5 AND user_id = $6
	`
	_, err := r.db.ExecContext(ctx, query, l.Name, l.Address, l.Phone, l.IsDefault, l.ID, l.UserID)
	return err
}

func (r *PostgresLocationRepository) Delete(ctx context.Context, id string, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM locations WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *PostgresLocationRepository) GetByID(ctx context.Context, id string, userID string) (*entity.Location, error) {
	var row locationRow
	err := r.db.GetContext(ctx, &row, `SELECT * FROM locations WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapLocationRowToEntity(row), nil
}

func (r *PostgresLocationRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Location, error) {
	var rows []locationRow
	err := r.db.SelectContext(ctx, &rows, `SELECT * FROM locations WHERE user_id = $1 ORDER BY created_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	locations := make([]*entity.Location, len(rows))
	for i, row := range rows {
		locations[i] = mapLocationRowToEntity(row)
	}
	return locations, nil
}

func (r *PostgresLocationRepository) GetDefault(ctx context.Context, userID string) (*entity.Location, error) {
	var row locationRow
	err := r.db.GetContext(ctx, &row, `SELECT * FROM locations WHERE user_id = $1 AND is_default = TRUE LIMIT 1`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapLocationRowToEntity(row), nil
}

func (r *PostgresLocationRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM locations WHERE user_id = $1`, userID)
	return count, err
}

func (r *PostgresLocationRepository) UnsetDefault(ctx context.Context, userID string, exceptID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE locations SET is_default = FALSE WHERE user_id = $1 AND id != $2`,
		userID, exceptID,
	)
	return err
}

func mapLocationRowToEntity(row locationRow) *entity.Location {
	return &entity.Location{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Address:   row.Address.String,
		Phone:     row.Phone.String,
		IsDefault: row.IsDefault,
		CreatedAt: row.CreatedAt,
	}
}
