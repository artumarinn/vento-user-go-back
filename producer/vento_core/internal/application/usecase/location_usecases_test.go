package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// ---- fakeLocationRepository ----

type fakeLocationRepository struct {
	locations map[string]*entity.Location
}

func newFakeLocationRepository() *fakeLocationRepository {
	return &fakeLocationRepository{locations: make(map[string]*entity.Location)}
}

func (f *fakeLocationRepository) Save(ctx context.Context, l *entity.Location) error {
	f.locations[l.ID] = l
	return nil
}
func (f *fakeLocationRepository) Update(ctx context.Context, l *entity.Location) error {
	f.locations[l.ID] = l
	return nil
}
func (f *fakeLocationRepository) Delete(ctx context.Context, id string, userID string) error {
	delete(f.locations, id)
	return nil
}
func (f *fakeLocationRepository) GetByID(ctx context.Context, id string, userID string) (*entity.Location, error) {
	l, ok := f.locations[id]
	if !ok || l.UserID != userID {
		return nil, nil
	}
	return l, nil
}
func (f *fakeLocationRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Location, error) {
	var res []*entity.Location
	for _, l := range f.locations {
		if l.UserID == userID {
			res = append(res, l)
		}
	}
	return res, nil
}
func (f *fakeLocationRepository) GetDefault(ctx context.Context, userID string) (*entity.Location, error) {
	for _, l := range f.locations {
		if l.UserID == userID && l.IsDefault {
			return l, nil
		}
	}
	return nil, nil
}
func (f *fakeLocationRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	count := 0
	for _, l := range f.locations {
		if l.UserID == userID {
			count++
		}
	}
	return count, nil
}
func (f *fakeLocationRepository) UnsetDefault(ctx context.Context, userID string, exceptID string) error {
	for _, l := range f.locations {
		if l.UserID == userID && l.ID != exceptID {
			l.IsDefault = false
		}
	}
	return nil
}

func TestCreateLocation_FirstLocationIsAlwaysDefault(t *testing.T) {
	repo := newFakeLocationRepository()
	uc := NewLocationUsecases(repo)

	res, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Sucursal Centro"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsDefault {
		t.Errorf("expected the first location to be default")
	}
}

func TestCreateLocation_SecondLocationNotDefaultUnlessRequested(t *testing.T) {
	repo := newFakeLocationRepository()
	uc := NewLocationUsecases(repo)

	_, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Local Central"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Sucursal Norte"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsDefault {
		t.Errorf("expected the second location to not be default by default")
	}
}

func TestCreateLocation_ExplicitDefaultUnsetsPreviousDefault(t *testing.T) {
	repo := newFakeLocationRepository()
	uc := NewLocationUsecases(repo)

	first, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Local Central"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Sucursal Norte", IsDefault: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updatedFirst, _ := repo.GetByID(context.Background(), first.ID, "user-1")
	if updatedFirst.IsDefault {
		t.Errorf("expected first location to no longer be default")
	}
}

func TestDeleteLocation_CannotDeleteDefaultLocation(t *testing.T) {
	repo := newFakeLocationRepository()
	uc := NewLocationUsecases(repo)

	first, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Local Central"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Sucursal Norte"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = uc.DeleteLocation(context.Background(), "user-1", first.ID)
	if !errors.Is(err, ErrCannotDeleteDefaultLocation) {
		t.Fatalf("expected ErrCannotDeleteDefaultLocation, got %v", err)
	}
}

func TestDeleteLocation_CannotDeleteLastLocation(t *testing.T) {
	repo := newFakeLocationRepository()
	uc := NewLocationUsecases(repo)

	only, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Local Central"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = uc.DeleteLocation(context.Background(), "user-1", only.ID)
	if !errors.Is(err, ErrCannotDeleteLastLocation) {
		t.Fatalf("expected ErrCannotDeleteLastLocation, got %v", err)
	}
}

func TestDeleteLocation_AllowsDeletingNonDefaultWhenMultipleExist(t *testing.T) {
	repo := newFakeLocationRepository()
	uc := NewLocationUsecases(repo)

	if _, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Local Central"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Sucursal Norte"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := uc.DeleteLocation(context.Background(), "user-1", second.ID); err != nil {
		t.Fatalf("unexpected error deleting non-default location: %v", err)
	}
}

func TestUpdateLocation_SettingDefaultUnsetsPreviousDefault(t *testing.T) {
	repo := newFakeLocationRepository()
	uc := NewLocationUsecases(repo)

	first, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Local Central"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := uc.CreateLocation(context.Background(), "user-1", dto.CreateLocationRequest{Name: "Sucursal Norte"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	makeDefault := true
	_, err = uc.UpdateLocation(context.Background(), "user-1", second.ID, dto.UpdateLocationRequest{IsDefault: &makeDefault})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updatedFirst, _ := repo.GetByID(context.Background(), first.ID, "user-1")
	if updatedFirst.IsDefault {
		t.Errorf("expected first location to no longer be default")
	}
	updatedSecond, _ := repo.GetByID(context.Background(), second.ID, "user-1")
	if !updatedSecond.IsDefault {
		t.Errorf("expected second location to be default")
	}
}

func TestGetOrCreateDefaultLocation_CreatesOneWhenUserHasNone(t *testing.T) {
	repo := newFakeLocationRepository()
	uc := NewLocationUsecases(repo)

	loc, err := uc.GetOrCreateDefaultLocation(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc == nil || !loc.IsDefault {
		t.Fatalf("expected a default location to be created, got %+v", loc)
	}

	count, _ := repo.CountByUserID(context.Background(), "user-1")
	if count != 1 {
		t.Errorf("expected exactly 1 location, got %d", count)
	}
}
