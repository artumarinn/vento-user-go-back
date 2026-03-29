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
	p := entity.NewProduct(userID, req.Name, req.SKU, req.Category, req.Price)
	p.Stock = req.Stock
	p.StockUnit = req.StockUnit
	p.MaxStock = req.MaxStock
	p.Supplier = req.Supplier
	p.SupplierCost = req.SupplierCost

	if err := uc.repo.Save(ctx, p); err != nil {
		return dto.ProductResponse{}, err
	}

	return mapEntityToDTO(p), nil
}

func (uc *CatalogUsecases) BatchCreateProducts(ctx context.Context, userID string, req dto.BatchCreateProductRequest) ([]dto.ProductResponse, error) {
	products := make([]*entity.Product, len(req.Products))
	res := make([]dto.ProductResponse, len(req.Products))

	for i, pReq := range req.Products {
		p := entity.NewProduct(userID, pReq.Name, pReq.SKU, pReq.Category, pReq.Price)
		p.Stock = pReq.Stock
		p.StockUnit = pReq.StockUnit
		p.MaxStock = pReq.MaxStock
		p.Supplier = pReq.Supplier
		p.SupplierCost = pReq.SupplierCost
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
		return dto.ProductResponse{}, nil // Should probably be an error
	}

	if req.Name != "" { p.Name = req.Name }
	if req.SKU != "" { p.SKU = req.SKU }
	if req.Category != "" { p.Category = req.Category }
	if req.Stock != 0 { p.Stock = req.Stock }
	if req.StockUnit != "" { p.StockUnit = req.StockUnit }
	if req.MaxStock != nil { p.MaxStock = req.MaxStock }
	if req.Price != 0 { p.Price = req.Price }
	if req.Supplier != "" { p.Supplier = req.Supplier }
	if req.SupplierCost != 0 { p.SupplierCost = req.SupplierCost }

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
		Stock:        p.Stock,
		StockUnit:    p.StockUnit,
		MaxStock:     p.MaxStock,
		Price:        p.Price,
		Supplier:     p.Supplier,
		SupplierCost: p.SupplierCost,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}
