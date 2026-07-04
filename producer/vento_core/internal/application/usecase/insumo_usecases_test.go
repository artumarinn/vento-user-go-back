package usecase

import (
	"context"
	"testing"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// TestListInsumos_IgnoresLocationNotOwnedByUser mirrors
// TestListProducts_IgnoresLocationNotOwnedByUser: a caller-supplied
// location_id that does not belong to the requesting user must never be
// used to read location_insumo_stock.
func TestListInsumos_IgnoresLocationNotOwnedByUser(t *testing.T) {
	insumoRepo := newFakeInsumoRepository()
	stock := 42.0
	insumoRepo.insumos["insumo-1"] = &entity.Insumo{ID: "insumo-1", UserID: "user-1", Name: "Harina", Stock: &stock}
	locationRepo := newFakeLocationRepository()
	otherTenantLocation := entity.NewLocation("user-2", "Sucursal Ajena", "", "", true)
	locationRepo.locations[otherTenantLocation.ID] = otherTenantLocation
	insumoRepo.locationStock = map[string]float64{otherTenantLocation.ID + "|insumo-1": 999}

	uc := NewInsumoUsecases(insumoRepo, &fakeTagRepository{}, locationRepo)

	res, err := uc.ListInsumos(context.Background(), "user-1", otherTenantLocation.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 insumo, got %d", len(res))
	}
	if res[0].Stock == nil || *res[0].Stock != 42 {
		t.Errorf("expected aggregate stock=42 (foreign location ignored), got %v", res[0].Stock)
	}
}

func TestListInsumos_UsesOwnedLocationStock(t *testing.T) {
	insumoRepo := newFakeInsumoRepository()
	stock := 42.0
	insumoRepo.insumos["insumo-1"] = &entity.Insumo{ID: "insumo-1", UserID: "user-1", Name: "Harina", Stock: &stock}
	locationRepo := newFakeLocationRepository()
	myLocation := entity.NewLocation("user-1", "Local Central", "", "", true)
	locationRepo.locations[myLocation.ID] = myLocation
	insumoRepo.locationStock = map[string]float64{myLocation.ID + "|insumo-1": 7}

	uc := NewInsumoUsecases(insumoRepo, &fakeTagRepository{}, locationRepo)

	res, err := uc.ListInsumos(context.Background(), "user-1", myLocation.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res[0].Stock == nil || *res[0].Stock != 7 {
		t.Errorf("expected location stock=7, got %v", res[0].Stock)
	}
}

func TestListInsumos_EmptyLocationReturnsAggregate(t *testing.T) {
	insumoRepo := newFakeInsumoRepository()
	stock := 42.0
	insumoRepo.insumos["insumo-1"] = &entity.Insumo{ID: "insumo-1", UserID: "user-1", Name: "Harina", Stock: &stock}
	locationRepo := newFakeLocationRepository()

	uc := NewInsumoUsecases(insumoRepo, &fakeTagRepository{}, locationRepo)

	res, err := uc.ListInsumos(context.Background(), "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res[0].Stock == nil || *res[0].Stock != 42 {
		t.Errorf("expected aggregate stock=42, got %v", res[0].Stock)
	}
}
