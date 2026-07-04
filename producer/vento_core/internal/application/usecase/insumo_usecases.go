package usecase

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type InsumoUsecases struct {
	repo         port.InsumoRepository
	tagRepo      port.TagRepository
	locationRepo port.LocationRepository
}

func NewInsumoUsecases(repo port.InsumoRepository, tagRepo port.TagRepository, locationRepo port.LocationRepository) *InsumoUsecases {
	return &InsumoUsecases{repo: repo, tagRepo: tagRepo, locationRepo: locationRepo}
}

// ListInsumos returns the insumo list for a user. When locationID is
// non-empty AND owned by userID, each insumo's Stock field is overridden
// with its per-location quantity (location_insumo_stock) instead of the
// shared aggregate — mirrors CatalogUsecases.ListProducts. A locationID that
// does not belong to userID is ignored and the aggregate view is returned
// instead, so a caller can never read another tenant's location_insumo_stock
// by guessing/reusing a foreign location UUID. Omitted locationID also keeps
// the aggregate view.
func (uc *InsumoUsecases) ListInsumos(ctx context.Context, userID string, locationID string) ([]dto.InsumoResponse, error) {
	insumos, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	ownedLocationID := ""
	if locationID != "" {
		loc, err := uc.locationRepo.GetByID(ctx, locationID, userID)
		if err != nil {
			return nil, err
		}
		if loc != nil {
			ownedLocationID = loc.ID
		}
	}

	res := make([]dto.InsumoResponse, len(insumos))
	for i, insumo := range insumos {
		tags, err := uc.tagRepo.ListByInsumoID(ctx, insumo.ID)
		if err != nil {
			return nil, err
		}
		res[i] = mapInsumoEntityToDTO(insumo, tags)
		if ownedLocationID != "" {
			stock, err := uc.repo.GetLocationInsumoStock(ctx, ownedLocationID, insumo.ID)
			if err != nil {
				return nil, err
			}
			res[i].Stock = ptr(stock)
		}
	}
	return res, nil
}

func (uc *InsumoUsecases) CreateInsumo(ctx context.Context, userID string, req dto.CreateInsumoRequest) (dto.InsumoResponse, error) {
	price := 0.0
	if req.Price != nil { price = *req.Price }

	i := entity.NewInsumo(userID, req.Name, req.Category, price)

	if req.Stock != nil { i.Stock = req.Stock }
	if req.StockUnit != "" { i.StockUnit = req.StockUnit }
	i.Supplier = req.Supplier

	if err := uc.repo.Save(ctx, i); err != nil {
		return dto.InsumoResponse{}, err
	}

	if err := uc.tagRepo.SetInsumoTags(ctx, i.ID, req.TagIDs); err != nil {
		return dto.InsumoResponse{}, err
	}

	tags, err := uc.tagRepo.ListByInsumoID(ctx, i.ID)
	if err != nil {
		return dto.InsumoResponse{}, err
	}

	return mapInsumoEntityToDTO(i, tags), nil
}

func (uc *InsumoUsecases) BatchCreateInsumos(ctx context.Context, userID string, req dto.BatchCreateInsumoRequest) ([]dto.InsumoResponse, error) {
	insumos := make([]*entity.Insumo, len(req.Insumos))
	res := make([]dto.InsumoResponse, len(req.Insumos))

	for i, iReq := range req.Insumos {
		price := 0.0
		if iReq.Price != nil { price = *iReq.Price }

		insumo := entity.NewInsumo(userID, iReq.Name, iReq.Category, price)

		if iReq.Stock != nil { insumo.Stock = iReq.Stock }
		if iReq.StockUnit != "" { insumo.StockUnit = iReq.StockUnit }
		insumo.Supplier = iReq.Supplier

		insumos[i] = insumo
		res[i] = mapInsumoEntityToDTO(insumo, nil)
	}

	if err := uc.repo.SaveBatch(ctx, insumos); err != nil {
		return nil, err
	}

	return res, nil
}

func (uc *InsumoUsecases) UpdateInsumo(ctx context.Context, userID string, insumoID string, req dto.UpdateInsumoRequest) (dto.InsumoResponse, error) {
	i, err := uc.repo.GetByID(ctx, insumoID, userID)
	if err != nil {
		return dto.InsumoResponse{}, err
	}
	if i == nil {
		return dto.InsumoResponse{}, nil
	}

	if req.Name != "" { i.Name = req.Name }
	if req.Category != "" { i.Category = req.Category }
	if req.Stock != nil { i.Stock = req.Stock }
	if req.StockUnit != "" { i.StockUnit = req.StockUnit }
	if req.Price != nil { i.Price = req.Price }
	if req.Supplier != "" { i.Supplier = req.Supplier }

	if err := uc.repo.Update(ctx, i); err != nil {
		return dto.InsumoResponse{}, err
	}

	if req.TagIDs != nil {
		if err := uc.tagRepo.SetInsumoTags(ctx, i.ID, req.TagIDs); err != nil {
			return dto.InsumoResponse{}, err
		}
	}

	tags, err := uc.tagRepo.ListByInsumoID(ctx, i.ID)
	if err != nil {
		return dto.InsumoResponse{}, err
	}

	return mapInsumoEntityToDTO(i, tags), nil
}

func (uc *InsumoUsecases) DeleteInsumo(ctx context.Context, userID string, insumoID string) error {
	return uc.repo.Delete(ctx, insumoID, userID)
}

func (uc *InsumoUsecases) GetInsumo(ctx context.Context, userID string, insumoID string) (*dto.InsumoResponse, error) {
	i, err := uc.repo.GetByID(ctx, insumoID, userID)
	if err != nil {
		return nil, err
	}
	if i == nil {
		return nil, nil
	}
	tags, err := uc.tagRepo.ListByInsumoID(ctx, i.ID)
	if err != nil {
		return nil, err
	}
	res := mapInsumoEntityToDTO(i, tags)
	return &res, nil
}

func mapInsumoEntityToDTO(i *entity.Insumo, tags []*entity.Tag) dto.InsumoResponse {
	tagResponses := make([]dto.TagResponse, len(tags))
	for idx, t := range tags {
		tagResponses[idx] = mapTagEntityToDTO(t)
	}

	return dto.InsumoResponse{
		ID:        i.ID,
		UserID:    i.UserID,
		Name:      i.Name,
		Category:  i.Category,
		Stock:     i.Stock,
		StockUnit: i.StockUnit,
		Price:     i.Price,
		Supplier:  i.Supplier,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Tags:      tagResponses,
	}
}
