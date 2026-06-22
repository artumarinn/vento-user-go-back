---
name: order-lifecycle
description: Estados de Pedido, tabla order_status_history y cómo cambiar el estado de una orden en vento-user-go-back. Usar al tocar órdenes, transiciones de estado o el historial de pedidos.
---

## Cuándo usar esta skill
- Cambiás el estado de una orden o agregás lógica sobre órdenes.
- Tocás el historial de estados o necesitás saber qué estados existen.

## Patrones establecidos

**Estados** (`internal/domain/entity/order.go`):
```go
type OrderStatus string
const (
	StatusPending    OrderStatus = "pending"
	StatusConfirmed  OrderStatus = "confirmed"
	StatusProcessing OrderStatus = "processing"
	StatusReady      OrderStatus = "ready"
	StatusDelivered  OrderStatus = "delivered"
	StatusPaused     OrderStatus = "paused"
	StatusCancelled  OrderStatus = "cancelled"
)
```

**Tabla de historial** (const en `infrastructure/adapter/postgres/connection.go`):
```sql
CREATE TABLE IF NOT EXISTS order_status_history (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Métodos del repo** (`infrastructure/adapter/postgres/order_repo.go`):
```go
func (r *PostgresOrderRepository) AppendStatusHistory(ctx context.Context, orderID string, status entity.OrderStatus) error {
	query := `INSERT INTO order_status_history (order_id, status, changed_at) VALUES ($1, $2, NOW())`
	_, err := r.db.ExecContext(ctx, query, orderID, string(status))
	return err
}
// GetStatusHistory: SELECT status, changed_at ... ORDER BY changed_at ASC → []entity.OrderStatusEvent
```

**Cambio de estado** (`internal/application/usecase/order_usecases.go`): el usecase valida contra un whitelist y delega en el repo, que llama `AppendStatusHistory()` automáticamente:
```go
func (uc *OrderUsecases) UpdateOrderStatus(ctx context.Context, userID, orderID, status string) (dto.OrderResponse, error) {
	if !validOrderStatuses[status] {
		return dto.OrderResponse{}, fmt.Errorf("invalid status: %s", status)
	}
	if err := uc.repo.UpdateStatus(ctx, orderID, userID, entity.OrderStatus(status)); err != nil {
		return dto.OrderResponse{}, err
	}
	// ... reload + map a DTO
}
```

## Convenciones de nombrado
- Constantes de estado: `Status<Nombre>` con valor en minúscula (`"pending"`).
- `validOrderStatuses` es un `map[string]bool` que lista los valores válidos.
- Evento de historial: `entity.OrderStatusEvent{Status, ChangedAt}`.

## Qué NO hacer
- NO cambies `orders.status` con un UPDATE crudo sin pasar por `UpdateStatus`/`AppendStatusHistory`: perdés la auditoría.
- NO asumas que hay máquina de estados: `validOrderStatuses` solo valida que el valor **existe**, NO valida transiciones ordenadas (cualquier estado puede ir a cualquier otro). Si necesitás reglas de transición (ej. no volver de `delivered` a `pending`), es trabajo nuevo a agregar explícitamente, no algo ya implementado.
- NO inventes estados fuera del enum: agregalos primero en `order.go` y en `validOrderStatuses`.

## Comandos útiles
```bash
cd producer/vento_core
go test ./internal/application/usecase -run TestOrder -v
go run cmd/server/main.go     # :8082
```
