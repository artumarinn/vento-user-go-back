package port

import (
	"context"
	"time"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type ClientStats struct {
	Client           *entity.Client
	TotalSpent       float64
	OrderCount       int
	LastPurchaseDate *time.Time
}

type ClientRepository interface {
	Save(ctx context.Context, client *entity.Client) error
	ListByUserIDWithStats(ctx context.Context, userID string) ([]ClientStats, error)
}
