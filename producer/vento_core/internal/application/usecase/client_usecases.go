package usecase

import (
	"context"
	"time"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type ClientUsecases struct {
	repo port.ClientRepository
}

func NewClientUsecases(repo port.ClientRepository) *ClientUsecases {
	return &ClientUsecases{repo: repo}
}

func (uc *ClientUsecases) CreateClient(ctx context.Context, userID string, req dto.CreateClientRequest) (dto.ClientResponse, error) {
	c := entity.NewClient(userID, req.Name, req.Phone, req.SocialNetwork, req.SocialHandle)
	if err := uc.repo.Save(ctx, c); err != nil {
		return dto.ClientResponse{}, err
	}
	return dto.ClientResponse{
		ID:                c.ID,
		Name:              c.Name,
		Phone:             c.Phone,
		SocialNetwork:     c.SocialNetwork,
		SocialHandle:      c.SocialHandle,
		LastPurchaseDate:  "",
		DaysSincePurchase: 9999,
		TotalSpent:        0,
		OrderCount:        0,
		Status:            "inactive",
	}, nil
}

func (uc *ClientUsecases) ListClients(ctx context.Context, userID string) ([]dto.ClientResponse, error) {
	stats, err := uc.repo.ListByUserIDWithStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.ClientResponse, len(stats))
	for i, s := range stats {
		var days int
		var status string
		var lastPurchaseStr string

		if s.LastPurchaseDate == nil {
			days = 9999
			status = "inactive"
			lastPurchaseStr = ""
		} else {
			days = int(time.Since(*s.LastPurchaseDate).Hours() / 24)
			switch {
			case days <= 30:
				status = "active"
			case days <= 90:
				status = "at_risk"
			default:
				status = "inactive"
			}
			lastPurchaseStr = s.LastPurchaseDate.Format(time.RFC3339)
		}

		res[i] = dto.ClientResponse{
			ID:                s.Client.ID,
			Name:              s.Client.Name,
			Phone:             s.Client.Phone,
			SocialNetwork:     s.Client.SocialNetwork,
			SocialHandle:      s.Client.SocialHandle,
			LastPurchaseDate:  lastPurchaseStr,
			DaysSincePurchase: days,
			TotalSpent:        s.TotalSpent,
			OrderCount:        s.OrderCount,
			Status:            status,
		}
	}
	return res, nil
}
