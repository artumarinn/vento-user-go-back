package usecase

import (
	"context"
	"errors"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// ErrCannotDeleteDefaultLocation is returned when a user tries to delete
// their default location. A different location must be promoted to default
// first.
var ErrCannotDeleteDefaultLocation = errors.New("location: cannot delete the default location")

// ErrCannotDeleteLastLocation is returned when a user tries to delete their
// only remaining location. Every user must have at least one location.
var ErrCannotDeleteLastLocation = errors.New("location: cannot delete the last remaining location")

type LocationUsecases struct {
	repo port.LocationRepository
}

func NewLocationUsecases(repo port.LocationRepository) *LocationUsecases {
	return &LocationUsecases{repo: repo}
}

func (uc *LocationUsecases) ListLocations(ctx context.Context, userID string) ([]dto.LocationResponse, error) {
	locations, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]dto.LocationResponse, len(locations))
	for i, l := range locations {
		res[i] = mapLocationEntityToDTO(l)
	}
	return res, nil
}

func (uc *LocationUsecases) CreateLocation(ctx context.Context, userID string, req dto.CreateLocationRequest) (dto.LocationResponse, error) {
	count, err := uc.repo.CountByUserID(ctx, userID)
	if err != nil {
		return dto.LocationResponse{}, err
	}

	// The very first location a user creates is always the default one —
	// there must never be a user with zero locations or zero defaults.
	isDefault := req.IsDefault || count == 0

	l := entity.NewLocation(userID, req.Name, req.Address, req.Phone, isDefault)

	if isDefault && count > 0 {
		if err := uc.repo.UnsetDefault(ctx, userID, ""); err != nil {
			return dto.LocationResponse{}, err
		}
	}

	if err := uc.repo.Save(ctx, l); err != nil {
		return dto.LocationResponse{}, err
	}

	return mapLocationEntityToDTO(l), nil
}

func (uc *LocationUsecases) UpdateLocation(ctx context.Context, userID string, locationID string, req dto.UpdateLocationRequest) (dto.LocationResponse, error) {
	l, err := uc.repo.GetByID(ctx, locationID, userID)
	if err != nil {
		return dto.LocationResponse{}, err
	}
	if l == nil {
		return dto.LocationResponse{}, nil
	}

	if req.Name != "" {
		l.Name = req.Name
	}
	if req.Address != "" {
		l.Address = req.Address
	}
	if req.Phone != "" {
		l.Phone = req.Phone
	}
	if req.IsDefault != nil && *req.IsDefault && !l.IsDefault {
		if err := uc.repo.UnsetDefault(ctx, userID, l.ID); err != nil {
			return dto.LocationResponse{}, err
		}
		l.IsDefault = true
	}

	if err := uc.repo.Update(ctx, l); err != nil {
		return dto.LocationResponse{}, err
	}

	return mapLocationEntityToDTO(l), nil
}

func (uc *LocationUsecases) DeleteLocation(ctx context.Context, userID string, locationID string) error {
	l, err := uc.repo.GetByID(ctx, locationID, userID)
	if err != nil {
		return err
	}
	if l == nil {
		return nil
	}

	count, err := uc.repo.CountByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrCannotDeleteLastLocation
	}
	if l.IsDefault {
		return ErrCannotDeleteDefaultLocation
	}

	return uc.repo.Delete(ctx, locationID, userID)
}

// GetOrCreateDefaultLocation returns the user's default location, creating
// "Local Central" on the fly if the user somehow has none (e.g. a user
// created before multi-location existed and not yet backfilled).
func (uc *LocationUsecases) GetOrCreateDefaultLocation(ctx context.Context, userID string) (*entity.Location, error) {
	existing, err := uc.repo.GetDefault(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	l := entity.NewLocation(userID, "Local Central", "", "", true)
	if err := uc.repo.Save(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func mapLocationEntityToDTO(l *entity.Location) dto.LocationResponse {
	return dto.LocationResponse{
		ID:        l.ID,
		UserID:    l.UserID,
		Name:      l.Name,
		Address:   l.Address,
		Phone:     l.Phone,
		IsDefault: l.IsDefault,
		CreatedAt: l.CreatedAt,
	}
}
