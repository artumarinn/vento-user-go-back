package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type productRow struct {
	ID           string      `db:"id"`
	UserID       string      `db:"user_id"`
	Name         string      `db:"name"`
	SKU          string      `db:"sku"`
	Category     string      `db:"category"`
	Stock        float64     `db:"stock"`
	StockUnit    string      `db:"stock_unit"`
	MaxStock     *float64    `db:"max_stock"`
	Price        float64     `db:"price"`
	Supplier     string      `db:"supplier"`
	SupplierCost float64     `db:"supplier_cost"`
	Metadata     interface{} `db:"metadata"`
	Description  string      `db:"description"`
	Tags         string      `db:"tags"`
	ImageURL     string      `db:"image_url"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
}

type PostgresProductRepository struct {
	db *sqlx.DB
}

func NewPostgresProductRepository(db *sqlx.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Save(ctx context.Context, p *entity.Product) error {
	query := `
		INSERT INTO products (id, user_id, name, sku, category, stock, stock_unit, max_stock, price, supplier, supplier_cost, description, tags, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (user_id, name) DO UPDATE SET
			sku         = EXCLUDED.sku,
			category    = EXCLUDED.category,
			stock       = EXCLUDED.stock,
			price       = EXCLUDED.price,
			supplier    = EXCLUDED.supplier,
			description = CASE WHEN EXCLUDED.description != '' THEN EXCLUDED.description ELSE products.description END,
			tags        = CASE WHEN EXCLUDED.tags != '' THEN EXCLUDED.tags ELSE products.tags END,
			image_url   = CASE WHEN EXCLUDED.image_url != '' THEN EXCLUDED.image_url ELSE products.image_url END,
			updated_at  = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.UserID, p.Name, p.SKU, p.Category, p.Stock, p.StockUnit, p.MaxStock,
		p.Price, p.Supplier, p.SupplierCost, p.Description, p.Tags, p.ImageURL, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *entity.Product) error {
	query := `
		UPDATE products
		SET name = $1, sku = $2, category = $3, stock = $4, stock_unit = $5, max_stock = $6,
		    price = $7, supplier = $8, supplier_cost = $9,
		    description = CASE WHEN $10 != '' THEN $10 ELSE description END,
		    tags = CASE WHEN $11 != '' THEN $11 ELSE tags END,
		    image_url = CASE WHEN $12 != '' THEN $12 ELSE image_url END,
		    updated_at = $13
		WHERE id = $14 AND user_id = $15
	`
	_, err := r.db.ExecContext(ctx, query,
		p.Name, p.SKU, p.Category, p.Stock, p.StockUnit, p.MaxStock,
		p.Price, p.Supplier, p.SupplierCost, p.Description, p.Tags, p.ImageURL,
		time.Now().UTC(), p.ID, p.UserID,
	)
	return err
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string, userID string) (*entity.Product, error) {
	var row productRow
	err := r.db.GetContext(ctx, &row, `SELECT * FROM products WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapRowToEntity(row), nil
}

func (r *PostgresProductRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Product, error) {
	var rows []productRow
	err := r.db.SelectContext(ctx, &rows, `SELECT * FROM products WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	products := make([]*entity.Product, len(rows))
	for i, row := range rows {
		products[i] = mapRowToEntity(row)
	}
	return products, nil
}

func (r *PostgresProductRepository) SaveBatch(ctx context.Context, products []*entity.Product) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO products (id, user_id, name, sku, category, stock, stock_unit, max_stock, price, supplier, supplier_cost, description, tags, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (user_id, name) DO UPDATE SET
			sku         = EXCLUDED.sku,
			category    = EXCLUDED.category,
			stock       = EXCLUDED.stock,
			price       = EXCLUDED.price,
			supplier    = EXCLUDED.supplier,
			description = CASE WHEN EXCLUDED.description != '' THEN EXCLUDED.description ELSE products.description END,
			tags        = CASE WHEN EXCLUDED.tags != '' THEN EXCLUDED.tags ELSE products.tags END,
			image_url   = CASE WHEN EXCLUDED.image_url != '' THEN EXCLUDED.image_url ELSE products.image_url END,
			updated_at  = EXCLUDED.updated_at
	`

	for _, p := range products {
		_, err := tx.ExecContext(ctx, query,
			p.ID, p.UserID, p.Name, p.SKU, p.Category, p.Stock, p.StockUnit, p.MaxStock,
			p.Price, p.Supplier, p.SupplierCost, p.Description, p.Tags, p.ImageURL, p.CreatedAt, p.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresProductRepository) Search(ctx context.Context, userID string, query string, category string, limit int) ([]*entity.Product, error) {
	args := []any{userID}
	where := []string{"user_id = $1"}
	idx := 2
	if q := strings.TrimSpace(query); q != "" {
		like := "%" + q + "%"
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", idx, idx))
		args = append(args, like)
		idx++
	}
	if category != "" {
		where = append(where, fmt.Sprintf("category = $%d", idx))
		args = append(args, category)
		idx++
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}

	sqlQuery := fmt.Sprintf(`SELECT * FROM products WHERE %s ORDER BY name ASC LIMIT $%d`,
		strings.Join(where, " AND "), idx)
	args = append(args, limit)

	var rows []productRow
	err := r.db.SelectContext(ctx, &rows, sqlQuery, args...)
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
		Description:  row.Description,
		Tags:         row.Tags,
		ImageURL:     row.ImageURL,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
