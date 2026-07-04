package usecase

import (
	"context"
	"strconv"
	"strings"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
)

type MetricsUsecases struct {
	repo port.MetricsRepository
}

func NewMetricsUsecases(repo port.MetricsRepository) *MetricsUsecases {
	return &MetricsUsecases{repo: repo}
}

func (uc *MetricsUsecases) GetMetrics(ctx context.Context, userID string, period string) (*dto.MetricsResponse, error) {
	days := parsePeriodDays(period)
	return uc.repo.FetchMetrics(ctx, userID, days)
}

func parsePeriodDays(period string) int {
	period = strings.ToLower(strings.TrimSuffix(period, "d"))
	n, err := strconv.Atoi(period)
	if err != nil || n <= 0 {
		return 7
	}
	if n > 365 {
		return 365
	}
	return n
}
