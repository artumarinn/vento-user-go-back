package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
	// unaccentAvailable is detected once at construction. When the unaccent
	// extension isn't installed (e.g. no superuser rights on a managed
	// Postgres), Search falls back to plain ILIKE with a manual accent-fold
	// applied to the query text instead of failing outright.
	unaccentAvailable bool
}

func NewPostgresProductRepository(db *sqlx.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db, unaccentAvailable: hasUnaccentExtension(db)}
}

func hasUnaccentExtension(db *sqlx.DB) bool {
	var exists bool
	if err := db.Get(&exists, `SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'unaccent')`); err != nil {
		return false
	}
	return exists
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
		// A malformed id (e.g. the LLM passing a product NAME instead of its
		// UUID) makes Postgres reject the query at the type level rather than
		// simply returning zero rows. Treat that the same as "not found" so
		// callers get a clean nil instead of a 500-worthy error.
		if errors.Is(err, sql.ErrNoRows) || isInvalidUUIDError(err) {
			return nil, nil
		}
		return nil, err
	}
	return mapRowToEntity(row), nil
}

// isInvalidUUIDError reports whether err is Postgres' 22P02
// (invalid_text_representation), raised when a non-UUID string is compared
// against a UUID column.
func isInvalidUUIDError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "22P02"
	}
	return false
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
		conditions, wordArgs, nextIdx := buildTextSearchConditions(q, r.unaccentAvailable, idx)
		if len(conditions) > 0 {
			where = append(where, conditions...)
			for _, a := range wordArgs {
				args = append(args, a)
			}
			idx = nextIdx
		}
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

// buildTextSearchConditions turns a free-text query into per-word ILIKE
// conditions ANDed together, so "llavero pla" still narrows results while
// "llaveros" (plural) still matches a product named "Llavero generico PLA".
//
// Matching is case-insensitive via ILIKE. When the unaccent extension is
// available it's also accent-insensitive on both sides of the comparison;
// otherwise we fall back to a manual accent-fold applied to the query word
// only (the DB column itself isn't normalized in that fallback path — a known,
// documented limitation of the dependency-free approach).
func buildTextSearchConditions(query string, unaccentAvailable bool, startIdx int) (conditions []string, args []string, nextIdx int) {
	idx := startIdx
	for _, word := range strings.Fields(query) {
		stem := stemSearchWord(word)
		if stem == "" {
			continue
		}
		pattern := "%" + stem + "%"
		if unaccentAvailable {
			conditions = append(conditions, fmt.Sprintf(
				"(unaccent(name) ILIKE unaccent($%d) OR unaccent(description) ILIKE unaccent($%d))",
				idx, idx,
			))
		} else {
			pattern = foldAccents(pattern)
			conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", idx, idx))
		}
		args = append(args, pattern)
		idx++
	}
	return conditions, args, idx
}

// stemSearchWord lowercases a query word and strips a single trailing 's' so
// plain plurals ("llaveros") match a singular product name ("Llavero").
// Deliberately no dependency on a stemming library — this is a cheap
// approximation, not linguistically exhaustive.
func stemSearchWord(word string) string {
	word = strings.ToLower(strings.TrimSpace(word))
	if len(word) > 1 && strings.HasSuffix(word, "s") {
		return strings.TrimSuffix(word, "s")
	}
	return word
}

var accentReplacer = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ñ", "n",
	"Á", "A", "À", "A", "Â", "A", "Ä", "A",
	"É", "E", "È", "E", "Ê", "E", "Ë", "E",
	"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
	"Ó", "O", "Ò", "O", "Ô", "O", "Ö", "O",
	"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
	"Ñ", "N",
)

// foldAccents is the dependency-free fallback used when the unaccent
// extension isn't installed on the target Postgres instance.
func foldAccents(s string) string {
	return accentReplacer.Replace(s)
}

func (r *PostgresProductRepository) AdjustStock(ctx context.Context, productID string, userID string, delta float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products SET stock = stock + $1, updated_at = $2 WHERE id = $3 AND user_id = $4`,
		delta, time.Now().UTC(), productID, userID,
	)
	return err
}

func (r *PostgresProductRepository) InsertStockMovement(ctx context.Context, movement *entity.StockMovement) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO stock_movements (id, user_id, product_id, location_id, order_id, quantity_delta, reason, created_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)`,
		movement.UserID, movement.ProductID, nullableUUID(movement.LocationID), movement.OrderID, movement.QuantityDelta, movement.Reason, movement.CreatedAt,
	)
	return err
}

func (r *PostgresProductRepository) AdjustLocationStock(ctx context.Context, locationID string, productID string, delta float64) error {
	if locationID == "" {
		return nil
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO location_stock (location_id, product_id, quantity, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (location_id, product_id) DO UPDATE SET
			quantity = location_stock.quantity + $3,
			updated_at = $4`,
		locationID, productID, delta, time.Now().UTC(),
	)
	return err
}

func (r *PostgresProductRepository) GetLocationStock(ctx context.Context, locationID string, productID string) (float64, error) {
	var quantity float64
	err := r.db.GetContext(ctx, &quantity,
		`SELECT quantity FROM location_stock WHERE location_id = $1 AND product_id = $2`,
		locationID, productID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return quantity, nil
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
