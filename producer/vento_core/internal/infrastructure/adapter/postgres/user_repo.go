package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type userRow struct {
	ID           string         `db:"id"`
	Email        string         `db:"email"`
	Password     string         `db:"password"`
	FullName     string         `db:"full_name"`
	BusinessName string         `db:"business_name"`
	GoogleID     sql.NullString `db:"google_id"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

type PostgresUserRepository struct {
	db *sqlx.DB
}

func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password, full_name, business_name, google_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	var googleID interface{}
	if user.GoogleID == "" {
		googleID = nil
	} else {
		googleID = user.GoogleID
	}

	_, err := r.db.ExecContext(ctx,
		query,
		user.ID,
		user.Email.String(),
		user.Password.Hash(),
		user.FullName,
		user.BusinessName,
		googleID,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var row userRow
	query := `SELECT * FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &row, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return entity.ReconstructUser(row.ID, row.Email, row.Password, row.FullName, row.BusinessName, row.GoogleID.String, row.CreatedAt, row.UpdatedAt), nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var row userRow
	query := `SELECT * FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return entity.ReconstructUser(row.ID, row.Email, row.Password, row.FullName, row.BusinessName, row.GoogleID.String, row.CreatedAt, row.UpdatedAt), nil
}

func (r *PostgresUserRepository) FindByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	var row userRow
	query := `SELECT * FROM users WHERE google_id = $1`
	err := r.db.GetContext(ctx, &row, query, googleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return entity.ReconstructUser(row.ID, row.Email, row.Password, row.FullName, row.BusinessName, row.GoogleID.String, row.CreatedAt, row.UpdatedAt), nil
}
