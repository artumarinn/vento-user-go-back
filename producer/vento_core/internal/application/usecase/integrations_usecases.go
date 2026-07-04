package usecase

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
)

// totalIntegrations is the number of channels we currently support surfacing
// in the dashboard, including MercadoPago which is not built yet.
const totalIntegrations = 4

type IntegrationsUsecases struct {
	metaRepo port.MetaRepository
}

func NewIntegrationsUsecases(metaRepo port.MetaRepository) *IntegrationsUsecases {
	return &IntegrationsUsecases{metaRepo: metaRepo}
}

// GetStatus reports which channels have a real, persisted meta_configs row
// for the user. It never invents connection state — MercadoPago is always
// false because that integration is not built yet.
func (uc *IntegrationsUsecases) GetStatus(ctx context.Context, userID string) (*dto.IntegrationsStatusResponse, error) {
	configs, err := uc.metaRepo.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	status := &dto.IntegrationsStatusResponse{
		MercadoPago: false,
		Total:       totalIntegrations,
	}

	for _, cfg := range configs {
		switch cfg.Channel {
		case "whatsapp":
			status.WhatsApp = true
		case "instagram":
			status.Instagram = true
		case "messenger":
			status.Facebook = true
		}
	}

	if status.WhatsApp {
		status.Connected++
	}
	if status.Instagram {
		status.Connected++
	}
	if status.Facebook {
		status.Connected++
	}

	return status, nil
}
