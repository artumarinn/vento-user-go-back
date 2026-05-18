package usecase

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type CatalogUsecases struct {
	repo port.ProductRepository
}

func NewCatalogUsecases(repo port.ProductRepository) *CatalogUsecases {
	return &CatalogUsecases{repo: repo}
}

func (uc *CatalogUsecases) ListProducts(ctx context.Context, userID string) ([]dto.ProductResponse, error) {
	products, err := uc.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		res[i] = mapEntityToDTO(p)
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

	if err := uc.repo.Save(ctx, p); err != nil {
		return dto.ProductResponse{}, err
	}

	return mapEntityToDTO(p), nil
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

		products[i] = p
		res[i] = mapEntityToDTO(p)
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

	if err := uc.repo.Update(ctx, p); err != nil {
		return dto.ProductResponse{}, err
	}

	return mapEntityToDTO(p), nil
}

func (uc *CatalogUsecases) DeleteProduct(ctx context.Context, userID string, productID string) error {
	return uc.repo.Delete(ctx, productID, userID)
}

func mapEntityToDTO(p *entity.Product) dto.ProductResponse {
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
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}
