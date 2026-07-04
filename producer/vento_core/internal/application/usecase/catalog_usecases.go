package usecase

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type CatalogUsecases struct {
	repo         port.ProductRepository
	tagRepo      port.TagRepository
	locationRepo port.LocationRepository
}

func NewCatalogUsecases(repo port.ProductRepository, tagRepo port.TagRepository, locationRepo port.LocationRepository) *CatalogUsecases {
	return &CatalogUsecases{repo: repo, tagRepo: tagRepo, locationRepo: locationRepo}
}

// ListProducts returns the catalog for a user. When locationID is non-empty
// AND owned by userID, each product's Stock field is overridden with its
// per-location quantity (location_stock) instead of the shared aggregate —
// used for the ?location_id= filtered inventory view. A locationID that
// does not belong to userID is ignored and the aggregate view is returned
// instead, so a caller can never read another tenant's location_stock by
// guessing/reusing a foreign location UUID. Omitted locationID also keeps
// the aggregate view.
func (uc *CatalogUsecases) ListProducts(ctx context.Context, userID string, locationID string) ([]dto.ProductResponse, error) {
	products, err := uc.repo.ListByUserID(ctx, userID)
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

	res := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		tags, err := uc.tagRepo.ListByProductID(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		res[i] = mapEntityToDTO(p, tags)
		if ownedLocationID != "" {
			stock, err := uc.repo.GetLocationStock(ctx, ownedLocationID, p.ID)
			if err != nil {
				return nil, err
			}
			res[i].Stock = ptr(stock)
		}
	}
	return res, nil
}

func (uc *CatalogUsecases) CreateProduct(ctx context.Context, userID string, req dto.CreateProductRequest) (dto.ProductResponse, error) {
	price := 0.0
	if req.Price != nil { price = *req.Price }

	p := entity.NewProduct(userID, req.Name, req.SKU, req.Category, price)

	if req.Stock != nil { p.Stock = *req.Stock }
	p.StockUnit = req.StockUnit
	p.MaxStock = req.MaxStock
	p.Supplier = req.Supplier
	if req.SupplierCost != nil { p.SupplierCost = *req.SupplierCost }
	p.Description = req.Description
	p.Tags = req.Tags
	p.ImageURL = req.ImageURL

	if err := uc.repo.Save(ctx, p); err != nil {
		return dto.ProductResponse{}, err
	}

	if err := uc.tagRepo.SetProductTags(ctx, p.ID, req.TagIDs); err != nil {
		return dto.ProductResponse{}, err
	}

	tags, err := uc.tagRepo.ListByProductID(ctx, p.ID)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapEntityToDTO(p, tags), nil
}

func (uc *CatalogUsecases) BatchCreateProducts(ctx context.Context, userID string, req dto.BatchCreateProductRequest) ([]dto.ProductResponse, error) {
	products := make([]*entity.Product, len(req.Products))
	res := make([]dto.ProductResponse, len(req.Products))

	for i, pReq := range req.Products {
		price := 0.0
		if pReq.Price != nil { price = *pReq.Price }

		p := entity.NewProduct(userID, pReq.Name, pReq.SKU, pReq.Category, price)

		if pReq.Stock != nil { p.Stock = *pReq.Stock }
		p.StockUnit = pReq.StockUnit
		if pReq.MaxStock != nil { p.MaxStock = pReq.MaxStock }
		p.Supplier = pReq.Supplier
		if pReq.SupplierCost != nil { p.SupplierCost = *pReq.SupplierCost }
		p.Description = pReq.Description
		p.Tags = pReq.Tags
		p.ImageURL = pReq.ImageURL

		products[i] = p
		res[i] = mapEntityToDTO(p, nil)
	}

	if err := uc.repo.SaveBatch(ctx, products); err != nil {
		return nil, err
	}

	return res, nil
}

func (uc *CatalogUsecases) UpdateProduct(ctx context.Context, userID string, productID string, req dto.UpdateProductRequest) (dto.ProductResponse, error) {
	p, err := uc.repo.GetByID(ctx, productID, userID)
	if err != nil {
		return dto.ProductResponse{}, err
	}
	if p == nil {
		return dto.ProductResponse{}, nil
	}

	if req.Name != "" { p.Name = req.Name }
	if req.SKU != "" { p.SKU = req.SKU }
	if req.Category != "" { p.Category = req.Category }
	if req.Stock != nil { p.Stock = *req.Stock }
	if req.StockUnit != "" { p.StockUnit = req.StockUnit }
	if req.MaxStock != nil { p.MaxStock = req.MaxStock }
	if req.Price != nil { p.Price = *req.Price }
	if req.Supplier != "" { p.Supplier = req.Supplier }
	if req.SupplierCost != nil { p.SupplierCost = *req.SupplierCost }
	if req.Description != "" { p.Description = req.Description }
	if req.Tags != "" { p.Tags = req.Tags }
	if req.ImageURL != "" { p.ImageURL = req.ImageURL }

	if err := uc.repo.Update(ctx, p); err != nil {
		return dto.ProductResponse{}, err
	}

	if req.TagIDs != nil {
		if err := uc.tagRepo.SetProductTags(ctx, p.ID, req.TagIDs); err != nil {
			return dto.ProductResponse{}, err
		}
	}

	tags, err := uc.tagRepo.ListByProductID(ctx, p.ID)
	if err != nil {
		return dto.ProductResponse{}, err
	}

	return mapEntityToDTO(p, tags), nil
}

func (uc *CatalogUsecases) DeleteProduct(ctx context.Context, userID string, productID string) error {
	return uc.repo.Delete(ctx, productID, userID)
}

func (uc *CatalogUsecases) GetProduct(ctx context.Context, userID string, productID string) (*dto.ProductResponse, error) {
	p, err := uc.repo.GetByID(ctx, productID, userID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	tags, err := uc.tagRepo.ListByProductID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	res := mapEntityToDTO(p, tags)
	return &res, nil
}

func mapEntityToDTO(p *entity.Product, tags []*entity.Tag) dto.ProductResponse {
	tagResponses := make([]dto.TagResponse, len(tags))
	for idx, t := range tags {
		tagResponses[idx] = mapTagEntityToDTO(t)
	}

	return dto.ProductResponse{
		ID:           p.ID,
		UserID:       p.UserID,
		Name:         p.Name,
		SKU:          p.SKU,
		Category:     p.Category,
		Stock:        ptr(p.Stock),
		StockUnit:    p.StockUnit,
		MaxStock:     p.MaxStock,
		Price:        ptr(p.Price),
		Supplier:     p.Supplier,
		SupplierCost: ptr(p.SupplierCost),
		Description:  p.Description,
		Tags:         p.Tags,
		ImageURL:     p.ImageURL,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		TagRefs:      tagResponses,
	}
}
