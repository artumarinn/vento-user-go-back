package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type SearchProductsUsecase struct {
	repo port.ProductRepository
}

func NewSearchProductsUsecase(repo port.ProductRepository) *SearchProductsUsecase {
	return &SearchProductsUsecase{repo: repo}
}

type SearchProductsInput struct {
	UserID     string
	Query      string
	MaxResults int
	Category   string
}

func (u *SearchProductsUsecase) Execute(ctx context.Context, in SearchProductsInput) ([]*entity.Product, error) {
	if strings.TrimSpace(in.Query) == "" && in.Category == "" {
		return nil, fmt.Errorf("query or category is required")
	}
	if in.MaxResults <= 0 || in.MaxResults > 50 {
		in.MaxResults = 20
	}
	return u.repo.Search(ctx, in.UserID, in.Query, in.Category, in.MaxResults)
}
