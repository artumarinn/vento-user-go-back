package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/shared/catalog"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type serviceRow struct {
	ID              string    `db:"id"`
	UserID          string    `db:"user_id"`
	Name            string    `db:"name"`
	Formula         string    `db:"formula"`
	MinimumLeadTime int       `db:"minimum_lead_time"`
	VariablesSchema []byte    `db:"variables_schema"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type PostgresServiceRepository struct {
	db *sqlx.DB
}

func NewPostgresServiceRepository(db *sqlx.DB) *PostgresServiceRepository {
	return &PostgresServiceRepository{db: db}
}

func (r *PostgresServiceRepository) Save(ctx context.Context, s *entity.Service) error {
	variablesJSON, _ := json.Marshal(s.VariablesSchema)
	query := `
		INSERT INTO services (id, user_id, name, formula, minimum_lead_time, variables_schema, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, name) DO UPDATE SET
			formula           = EXCLUDED.formula,
			minimum_lead_time = EXCLUDED.minimum_lead_time,
			variables_schema  = EXCLUDED.variables_schema,
			updated_at        = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.UserID, s.Name, s.Formula, s.MinimumLeadTime, variablesJSON, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *PostgresServiceRepository) Update(ctx context.Context, s *entity.Service) error {
	variablesJSON, _ := json.Marshal(s.VariablesSchema)
	query := `
		UPDATE services
		SET name = $1, formula = $2, minimum_lead_time = $3, variables_schema = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7
	`
	_, err := r.db.ExecContext(ctx, query,
		s.Name, s.Formula, s.MinimumLeadTime, variablesJSON, time.Now().UTC(), s.ID, s.UserID,
	)
	return err
}

func (r *PostgresServiceRepository) Delete(ctx context.Context, id string, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM services WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *PostgresServiceRepository) GetByID(ctx context.Context, id string, userID string) (*entity.Service, error) {
	var row serviceRow
	err := r.db.GetContext(ctx, &row, `SELECT id, user_id, name, formula, minimum_lead_time, variables_schema, created_at, updated_at FROM services WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapServiceRowToEntity(row), nil
}

func (r *PostgresServiceRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Service, error) {
	var rows []serviceRow
	err := r.db.SelectContext(ctx, &rows, `SELECT id, user_id, name, formula, minimum_lead_time, variables_schema, created_at, updated_at FROM services WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	services := make([]*entity.Service, len(rows))
	for i, row := range rows {
		services[i] = mapServiceRowToEntity(row)
	}
	return services, nil
}

func (r *PostgresServiceRepository) SaveBatch(ctx context.Context, services []*entity.Service) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO services (id, user_id, name, formula, minimum_lead_time, variables_schema, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, name) DO UPDATE SET
			formula           = EXCLUDED.formula,
			minimum_lead_time = EXCLUDED.minimum_lead_time,
			variables_schema  = EXCLUDED.variables_schema,
			updated_at        = EXCLUDED.updated_at
	`

	for _, s := range services {
		variablesJSON, _ := json.Marshal(s.VariablesSchema)
		_, err := tx.ExecContext(ctx, query,
			s.ID, s.UserID, s.Name, s.Formula, s.MinimumLeadTime, variablesJSON, s.CreatedAt, s.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func mapServiceRowToEntity(row serviceRow) *entity.Service {
	var variablesSchema []catalog.VariableDefinition
	_ = json.Unmarshal(row.VariablesSchema, &variablesSchema)
	return &entity.Service{
		ID:              row.ID,
		UserID:          row.UserID,
		Name:            row.Name,
		Formula:         row.Formula,
		MinimumLeadTime: row.MinimumLeadTime,
		VariablesSchema: variablesSchema,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
