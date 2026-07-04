package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/http/handler"
)

type fakeOrderRepo struct {
	saveFunc                func(ctx context.Context, order *entity.Order) error
	getByIDFunc             func(ctx context.Context, id string) (*entity.Order, error)
	getByIDForUserFunc      func(ctx context.Context, id string, userID string) (*entity.Order, error)
	listByUserIDFunc        func(ctx context.Context, userID string) ([]*entity.Order, error)
	updateStatusFunc        func(ctx context.Context, id string, userID string, status entity.OrderStatus) error
	saveBatchFunc           func(ctx context.Context, orders []*entity.Order) error
	updateFunc              func(ctx context.Context, order *entity.Order) error
	deleteFunc              func(ctx context.Context, id, userID string) error
	appendStatusHistoryFunc func(ctx context.Context, orderID string, status entity.OrderStatus) error
	getStatusHistoryFunc    func(ctx context.Context, orderID string) ([]entity.OrderStatusEvent, error)
	registerPaymentFunc     func(ctx context.Context, orderID string, method string, amount float64) error
}

func (f *fakeOrderRepo) Save(ctx context.Context, o *entity.Order) error { return f.saveFunc(ctx, o) }
func (f *fakeOrderRepo) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	return f.getByIDFunc(ctx, id)
}
func (f *fakeOrderRepo) GetByIDForUser(ctx context.Context, id string, userID string) (*entity.Order, error) {
	return f.getByIDForUserFunc(ctx, id, userID)
}
func (f *fakeOrderRepo) ListByUserID(ctx context.Context, u string) ([]*entity.Order, error) {
	return f.listByUserIDFunc(ctx, u)
}
func (f *fakeOrderRepo) ListByUserIDAndLocation(ctx context.Context, u string, locationID string) ([]*entity.Order, error) {
	return f.listByUserIDFunc(ctx, u)
}
func (f *fakeOrderRepo) UpdateStatus(ctx context.Context, id string, userID string, s entity.OrderStatus) error {
	return f.updateStatusFunc(ctx, id, userID, s)
}
func (f *fakeOrderRepo) SaveBatch(ctx context.Context, o []*entity.Order) error {
	return f.saveBatchFunc(ctx, o)
}
func (f *fakeOrderRepo) Update(ctx context.Context, o *entity.Order) error {
	return f.updateFunc(ctx, o)
}
func (f *fakeOrderRepo) Delete(ctx context.Context, id, userID string) error {
	return f.deleteFunc(ctx, id, userID)
}
func (f *fakeOrderRepo) AppendStatusHistory(ctx context.Context, orderID string, status entity.OrderStatus) error {
	return f.appendStatusHistoryFunc(ctx, orderID, status)
}
func (f *fakeOrderRepo) GetStatusHistory(ctx context.Context, orderID string) ([]entity.OrderStatusEvent, error) {
	return f.getStatusHistoryFunc(ctx, orderID)
}
func (f *fakeOrderRepo) RegisterPayment(ctx context.Context, orderID string, method string, amount float64) error {
	return f.registerPaymentFunc(ctx, orderID, method, amount)
}
func (f *fakeOrderRepo) InsertOrderPayment(ctx context.Context, payment *entity.OrderPayment) error {
	return nil
}
func (f *fakeOrderRepo) ListOrderPayments(ctx context.Context, orderID string) ([]entity.OrderPayment, error) {
	return nil, nil
}
func (f *fakeOrderRepo) ListOrderPaymentsByUserID(ctx context.Context, userID string) ([]entity.OrderPaymentDetail, error) {
	return nil, nil
}
func (f *fakeOrderRepo) SumPaidOrderPayments(ctx context.Context, orderID string) (float64, error) {
	return 0, nil
}
func (f *fakeOrderRepo) UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus string) error {
	return nil
}

type fakeServiceRepo struct{}

func (f *fakeServiceRepo) Save(ctx context.Context, service *entity.Service) error   { return nil }
func (f *fakeServiceRepo) Update(ctx context.Context, service *entity.Service) error { return nil }
func (f *fakeServiceRepo) Delete(ctx context.Context, id string, userID string) error {
	return nil
}
func (f *fakeServiceRepo) GetByID(ctx context.Context, id string, userID string) (*entity.Service, error) {
	return nil, nil
}
func (f *fakeServiceRepo) ListByUserID(ctx context.Context, userID string) ([]*entity.Service, error) {
	return nil, nil
}
func (f *fakeServiceRepo) SaveBatch(ctx context.Context, services []*entity.Service) error {
	return nil
}
func (f *fakeServiceRepo) SetServiceInsumos(ctx context.Context, serviceID string, insumos []entity.ServiceInsumo) error {
	return nil
}
func (f *fakeServiceRepo) ListServiceInsumos(ctx context.Context, serviceID string) ([]entity.ServiceInsumoDetail, error) {
	return nil, nil
}

type fakeInsumoRepo struct {
	listByUserIDFunc func(ctx context.Context, userID string) ([]*entity.Insumo, error)
}

func (f *fakeInsumoRepo) Save(ctx context.Context, insumo *entity.Insumo) error   { return nil }
func (f *fakeInsumoRepo) Update(ctx context.Context, insumo *entity.Insumo) error { return nil }
func (f *fakeInsumoRepo) Delete(ctx context.Context, id string, userID string) error {
	return nil
}
func (f *fakeInsumoRepo) GetByID(ctx context.Context, id string, userID string) (*entity.Insumo, error) {
	return nil, nil
}
func (f *fakeInsumoRepo) ListByUserID(ctx context.Context, userID string) ([]*entity.Insumo, error) {
	if f.listByUserIDFunc != nil {
		return f.listByUserIDFunc(ctx, userID)
	}
	return nil, nil
}
func (f *fakeInsumoRepo) SaveBatch(ctx context.Context, insumos []*entity.Insumo) error {
	return nil
}
func (f *fakeInsumoRepo) AdjustStock(ctx context.Context, insumoID string, userID string, delta float64) error {
	return nil
}
func (f *fakeInsumoRepo) InsertInsumoMovement(ctx context.Context, movement *entity.InsumoMovement) error {
	return nil
}
func (f *fakeInsumoRepo) AdjustLocationInsumoStock(ctx context.Context, locationID string, insumoID string, delta float64) error {
	return nil
}
func (f *fakeInsumoRepo) GetLocationInsumoStock(ctx context.Context, locationID string, insumoID string) (float64, error) {
	return 0, nil
}

type fakeLocationRepo struct{}

func (f *fakeLocationRepo) Save(ctx context.Context, l *entity.Location) error   { return nil }
func (f *fakeLocationRepo) Update(ctx context.Context, l *entity.Location) error { return nil }
func (f *fakeLocationRepo) Delete(ctx context.Context, id string, userID string) error {
	return nil
}
func (f *fakeLocationRepo) GetByID(ctx context.Context, id string, userID string) (*entity.Location, error) {
	return nil, nil
}
func (f *fakeLocationRepo) ListByUserID(ctx context.Context, userID string) ([]*entity.Location, error) {
	return nil, nil
}
func (f *fakeLocationRepo) GetDefault(ctx context.Context, userID string) (*entity.Location, error) {
	return &entity.Location{ID: "loc-default", UserID: userID, Name: "Local Central", IsDefault: true}, nil
}
func (f *fakeLocationRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	return 1, nil
}
func (f *fakeLocationRepo) UnsetDefault(ctx context.Context, userID string, exceptID string) error {
	return nil
}

func TestOrderHandler_List(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		repo := &fakeOrderRepo{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]*entity.Order, error) {
				return []*entity.Order{}, nil
			},
		}
		uc := usecase.NewOrderUsecases(repo, &fakeServiceRepo{}, &fakeProductRepo{}, &fakeInsumoRepo{}, &fakeLocationRepo{})
		locationUC := usecase.NewLocationUsecases(&fakeLocationRepo{})
		h := handler.NewOrderHandler(uc, locationUC)
		r, v1 := setupTestRouter()
		v1.GET("/orders", func(c *gin.Context) {
			c.Set("userID", "user-123")
			h.List(c)
		})

		w := performRequest(r, "GET", "/api/v1/orders", nil)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
