package port

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
)

type MetricsRepository interface {
	FetchMetrics(ctx context.Context, userID string, days int) (*dto.MetricsResponse, error)
}
