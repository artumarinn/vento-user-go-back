package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type insumoRow struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Name      string    `db:"name"`
	Category  string    `db:"category"`
	Stock     float64   `db:"stock"`
	StockUnit string    `db:"stock_unit"`
	Price     float64   `db:"price"`
	Supplier  string    `db:"supplier"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PostgresInsumoRepository struct {
	db *sqlx.DB
}

func NewPostgresInsumoRepository(db *sqlx.DB) *PostgresInsumoRepository {
	return &PostgresInsumoRepository{db: db}
}

func (r *PostgresInsumoRepository) Save(ctx context.Context, i *entity.Insumo) error {
	query := `
		INSERT INTO insumos (id, user_id, name, category, stock, stock_unit, price, supplier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, name) DO UPDATE SET
			category   = EXCLUDED.category,
			stock      = EXCLUDED.stock,
			price      = EXCLUDED.price,
			supplier   = EXCLUDED.supplier,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		i.ID, i.UserID, i.Name, i.Category, i.Stock, i.StockUnit,
		i.Price, i.Supplier, i.CreatedAt, i.UpdatedAt,
	)
	return err
}

func (r *PostgresInsumoRepository) Update(ctx context.Context, i *entity.Insumo) error {
	query := `
		UPDATE insumos
		SET name = $1, category = $2, stock = $3, stock_unit = $4,
		    price = $5, supplier = $6, updated_at = $7
		WHERE id = $8 AND user_id = $9
	`
	_, err := r.db.ExecContext(ctx, query,
		i.Name, i.Category, i.Stock, i.StockUnit,
		i.Price, i.Supplier, time.Now().UTC(), i.ID, i.UserID,
	)
	return err
}

func (r *PostgresInsumoRepository) Delete(ctx context.Context, id string, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM insumos WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *PostgresInsumoRepository) GetByID(ctx context.Context, id string, userID string) (*entity.Insumo, error) {
	var row insumoRow
	err := r.db.GetContext(ctx, &row, `SELECT * FROM insumos WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapInsumoRowToEntity(row), nil
}

func (r *PostgresInsumoRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Insumo, error) {
	var rows []insumoRow
	err := r.db.SelectContext(ctx, &rows, `SELECT * FROM insumos WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	insumos := make([]*entity.Insumo, len(rows))
	for i, row := range rows {
		insumos[i] = mapInsumoRowToEntity(row)
	}
	return insumos, nil
}

func (r *PostgresInsumoRepository) SaveBatch(ctx context.Context, insumos []*entity.Insumo) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO insumos (id, user_id, name, category, stock, stock_unit, price, supplier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, name) DO UPDATE SET
			category   = EXCLUDED.category,
			stock      = EXCLUDED.stock,
			price      = EXCLUDED.price,
			supplier   = EXCLUDED.supplier,
			updated_at = EXCLUDED.updated_at
	`

	for _, i := range insumos {
		_, err := tx.ExecContext(ctx, query,
			i.ID, i.UserID, i.Name, i.Category, i.Stock, i.StockUnit,
			i.Price, i.Supplier, i.CreatedAt, i.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresInsumoRepository) AdjustStock(ctx context.Context, insumoID string, userID string, delta float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE insumos SET stock = COALESCE(stock, 0) + $1, updated_at = $2 WHERE id = $3 AND user_id = $4`,
		delta, time.Now().UTC(), insumoID, userID,
	)
	return err
}

func (r *PostgresInsumoRepository) InsertInsumoMovement(ctx context.Context, movement *entity.InsumoMovement) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO insumo_movements (id, user_id, insumo_id, location_id, order_id, quantity_delta, reason, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)`,
		movement.UserID, movement.InsumoID, nullableUUID(movement.LocationID), movement.OrderID, movement.QuantityDelta, movement.Reason, movement.CreatedAt,
	)
	return err
}

func (r *PostgresInsumoRepository) AdjustLocationInsumoStock(ctx context.Context, locationID string, insumoID string, delta float64) error {
	if locationID == "" {
		return nil
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO location_insumo_stock (location_id, insumo_id, quantity, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (location_id, insumo_id) DO UPDATE SET
			quantity = location_insumo_stock.quantity + $3,
			updated_at = $4`,
		locationID, insumoID, delta, time.Now().UTC(),
	)
	return err
}

func (r *PostgresInsumoRepository) GetLocationInsumoStock(ctx context.Context, locationID string, insumoID string) (float64, error) {
	var quantity float64
	err := r.db.GetContext(ctx, &quantity,
		`SELECT quantity FROM location_insumo_stock WHERE location_id = $1 AND insumo_id = $2`,
		locationID, insumoID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return quantity, nil
}

func mapInsumoRowToEntity(row insumoRow) *entity.Insumo {
	stock := row.Stock
	price := row.Price
	return &entity.Insumo{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Category:  row.Category,
		Stock:     &stock,
		StockUnit: row.StockUnit,
		Price:     &price,
		Supplier:  row.Supplier,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
