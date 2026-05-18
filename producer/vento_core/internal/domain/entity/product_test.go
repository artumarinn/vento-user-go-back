package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

func TestNewProduct(t *testing.T) {
	p := entity.NewProduct("user-1", "Remera", "SKU-001", "Indumentaria", 1500.0)

	assert.NotEmpty(t, p.ID)
	assert.Equal(t, "user-1", p.UserID)
	assert.Equal(t, "Remera", p.Name)
	assert.Equal(t, "SKU-001", p.SKU)
	assert.Equal(t, "Indumentaria", p.Category)
	assert.Equal(t, 1500.0, p.Price)
	assert.Equal(t, "u", p.StockUnit)
	assert.False(t, p.CreatedAt.IsZero())
	assert.False(t, p.UpdatedAt.IsZero())
}

func TestNewProduct_UniqueIDs(t *testing.T) {
	p1 := entity.NewProduct("u", "A", "", "", 0)
	p2 := entity.NewProduct("u", "B", "", "", 0)
	assert.NotEqual(t, p1.ID, p2.ID)
}
