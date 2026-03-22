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

// userRow maps to the users table row.
type userRow struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	FullName  string    `db:"full_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// PostgresUserRepository implements UserRepository port with PostgreSQL.
type PostgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository.
func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Save inserts a new user into the database.
func (r *PostgresUserRepository) Save(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password, full_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx,
		query,
		user.ID,
		user.Email.String(),
		user.Password.Hash(),
		user.FullName,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

// FindByEmail retrieves a user by email address.
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var row userRow
	query := `SELECT id, email, password, full_name, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &row, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return entity.ReconstructUser(row.ID, row.Email, row.Password, row.FullName, row.CreatedAt, row.UpdatedAt), nil
}

// FindByID retrieves a user by ID.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var row userRow
	query := `SELECT id, email, password, full_name, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return entity.ReconstructUser(row.ID, row.Email, row.Password, row.FullName, row.CreatedAt, row.UpdatedAt), nil
}
