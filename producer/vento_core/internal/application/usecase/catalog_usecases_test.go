package usecase

import (
	"context"
	"testing"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// ---- fakeTagRepository ----

type fakeTagRepository struct{}

func (f *fakeTagRepository) Save(ctx context.Context, tag *entity.Tag) error { return nil }
func (f *fakeTagRepository) Delete(ctx context.Context, id string, userID string) error {
	return nil
}
func (f *fakeTagRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Tag, error) {
	return nil, nil
}
func (f *fakeTagRepository) ListByInsumoID(ctx context.Context, insumoID string) ([]*entity.Tag, error) {
	return nil, nil
}
func (f *fakeTagRepository) SetInsumoTags(ctx context.Context, insumoID string, tagIDs []string) error {
	return nil
}
func (f *fakeTagRepository) ListByProductID(ctx context.Context, productID string) ([]*entity.Tag, error) {
	return nil, nil
}
func (f *fakeTagRepository) SetProductTags(ctx context.Context, productID string, tagIDs []string) error {
	return nil
}
func (f *fakeTagRepository) ListByServiceID(ctx context.Context, serviceID string) ([]*entity.Tag, error) {
	return nil, nil
}
func (f *fakeTagRepository) SetServiceTags(ctx context.Context, serviceID string, tagIDs []string) error {
	return nil
}

// TestListProducts_IgnoresLocationNotOwnedByUser guards against a defense-
// in-depth gap: a caller-supplied location_id that does not belong to the
// requesting user must never be used to read location_stock — the aggregate
// view is returned instead of leaking/using another tenant's location.
func TestListProducts_IgnoresLocationNotOwnedByUser(t *testing.T) {
	productRepo := newFakeProductRepository()
	productRepo.products["prod-1"] = &entity.Product{ID: "prod-1", UserID: "user-1", Name: "Widget", Stock: 42}
	locationRepo := newFakeLocationRepository()
	otherTenantLocation := entity.NewLocation("user-2", "Sucursal Ajena", "", "", true)
	locationRepo.locations[otherTenantLocation.ID] = otherTenantLocation
	// Seed a location_stock value at that foreign location to prove it's
	// never read.
	productRepo.locationStock = map[string]float64{otherTenantLocation.ID + "|prod-1": 999}

	uc := NewCatalogUsecases(productRepo, &fakeTagRepository{}, locationRepo)

	res, err := uc.ListProducts(context.Background(), "user-1", otherTenantLocation.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 product, got %d", len(res))
	}
	if res[0].Stock == nil || *res[0].Stock != 42 {
		t.Errorf("expected aggregate stock=42 (foreign location ignored), got %v", res[0].Stock)
	}
}

func TestListProducts_UsesOwnedLocationStock(t *testing.T) {
	productRepo := newFakeProductRepository()
	productRepo.products["prod-1"] = &entity.Product{ID: "prod-1", UserID: "user-1", Name: "Widget", Stock: 42}
	locationRepo := newFakeLocationRepository()
	myLocation := entity.NewLocation("user-1", "Local Central", "", "", true)
	locationRepo.locations[myLocation.ID] = myLocation
	productRepo.locationStock = map[string]float64{myLocation.ID + "|prod-1": 7}

	uc := NewCatalogUsecases(productRepo, &fakeTagRepository{}, locationRepo)

	res, err := uc.ListProducts(context.Background(), "user-1", myLocation.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res[0].Stock == nil || *res[0].Stock != 7 {
		t.Errorf("expected location stock=7, got %v", res[0].Stock)
	}
}

// TestGetProduct_NonExistentID_ReturnsNilNotError guards the 404-not-500
// contract: GetProduct (backing ToolHandler.GetStock/GetProductDetails) must
// return (nil, nil) for an id that doesn't resolve to a product, so the HTTP
// handler can answer 404 instead of 500. The malformed-UUID case is covered
// at the postgres repo layer (product_repo_integration_test.go), since a
// fake repo can't reproduce Postgres' 22P02 type error.
func TestGetProduct_NonExistentID_ReturnsNilNotError(t *testing.T) {
	productRepo := newFakeProductRepository()
	uc := NewCatalogUsecases(productRepo, &fakeTagRepository{}, newFakeLocationRepository())

	res, err := uc.GetProduct(context.Background(), "user-1", "does-not-exist")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil product for non-existent id, got %+v", res)
	}
}
