package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type productRow struct {
	ID           string    `db:"id"`
	UserID       string    `db:"user_id"`
	Name         string    `db:"name"`
	SKU          string    `db:"sku"`
	Category     string    `db:"category"`
	Stock        float64   `db:"stock"`
	StockUnit    string    `db:"stock_unit"`
	MaxStock     *float64  `db:"max_stock"`
	Price        float64   `db:"price"`
	Supplier     string    `db:"supplier"`
	SupplierCost float64   `db:"supplier_cost"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type PostgresProductRepository struct {
	db *sqlx.DB
}

func NewPostgresProductRepository(db *sqlx.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Save(ctx context.Context, p *entity.Product) error {
	query := `
		INSERT INTO products (id, user_id, name, sku, category, stock, stock_unit, max_stock, price, supplier, supplier_cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.UserID, p.Name, p.SKU, p.Category, p.Stock, p.StockUnit, p.MaxStock, p.Price, p.Supplier, p.SupplierCost, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *entity.Product) error {
	query := `
		UPDATE products
		SET name = $1, sku = $2, category = $3, stock = $4, stock_unit = $5, max_stock = $6, price = $7, supplier = $8, supplier_cost = $9, updated_at = $10
		WHERE id = $11 AND user_id = $12
	`
	_, err := r.db.ExecContext(ctx, query,
		p.Name, p.SKU, p.Category, p.Stock, p.StockUnit, p.MaxStock, p.Price, p.Supplier, p.SupplierCost, time.Now().UTC(), p.ID, p.UserID,
	)
	return err
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string, userID string) error {
	query := `DELETE FROM products WHERE id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string, userID string) (*entity.Product, error) {
	var row productRow
	query := `SELECT * FROM products WHERE id = $1 AND user_id = $2`
	err := r.db.GetContext(ctx, &row, query, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Or a specific ErrProductNotFound
		}
		return nil, err
	}
	return mapRowToEntity(row), nil
}

func (r *PostgresProductRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Product, error) {
	var rows []productRow
	query := `SELECT * FROM products WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &rows, query, userID)
	if err != nil {
		return nil, err
	}

	products := make([]*entity.Product, len(rows))
	for i, row := range rows {
		products[i] = mapRowToEntity(row)
	}
	return products, nil
}

func mapRowToEntity(row productRow) *entity.Product {
	return &entity.Product{
		ID:           row.ID,
		UserID:       row.UserID,
		Name:         row.Name,
		SKU:          row.SKU,
		Category:     row.Category,
		Stock:        row.Stock,
		StockUnit:    row.StockUnit,
		MaxStock:     row.MaxStock,
		Price:        row.Price,
		Supplier:     row.Supplier,
		SupplierCost: row.SupplierCost,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
