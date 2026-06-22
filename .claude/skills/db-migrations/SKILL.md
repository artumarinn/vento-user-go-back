---
name: db-migrations
description: Cómo se crean y nombran migraciones en vento-user-go-back y las convenciones de schema (UUID PK, FK a users, timestamps, trigger updated_at). Usar al agregar tablas o columnas en el Core.
---

## Cuándo usar esta skill
- Agregás una tabla o columna nueva al Core.
- Necesitás entender cómo/dónde se aplican las migraciones al arrancar.

## Patrones establecidos

**Dos lugares con SQL — IMPORTANTE, verificá cuál aplica tu cambio:**
1. Archivos `migrations/NNN_descripcion.sql` (numeración de 3 dígitos, secuencial):
   ```
   001_create_users.sql ... 013_create_insumos.sql 014_create_tags.sql
   015_create_product_tags.sql 016_create_services.sql 017_create_service_tags.sql
   ```
2. **Const SQL embebido en `internal/infrastructure/adapter/postgres/connection.go`**, ejecutado en `NewConnection()` al startup:
   ```go
   migrations := []struct{ name, sql string }{
       {"users", createUsersTable}, {"products", createProductsTable},
       {"orders", createOrdersTable}, {"order_status_history", createOrderStatusHistoryTable},
       {"clients", createClientsTable}, ...
   }
   for _, m := range migrations {
       if _, err := db.Exec(m.sql); err != nil { return nil, fmt.Errorf("migration %q failed: %w", m.name, err) }
   }
   ```
   Las const (`createUsersTable`, etc.) usan `CREATE TABLE IF NOT EXISTS` → idempotentes.

   **No está 100% claro cuál de los dos es la fuente ejecutada para tablas nuevas como `services`/`insumos`**: el array visible en `connection.go` no incluye services/insumos/tags, pero los `.sql` correspondientes sí existen. Antes de agregar una migración, leé `connection.go` completo y confirmá si tu tabla se crea ahí o solo como `.sql`. Seguí el mecanismo que ya use la tabla más parecida.

**Convención de schema** (de `016_create_services.sql`):
```sql
CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    formula TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);
CREATE INDEX IF NOT EXISTS idx_services_user_id ON services(user_id);
CREATE TRIGGER update_services_updated_at
BEFORE UPDATE ON services FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

## Convenciones de nombrado
- Archivo: `NNN_create_<tabla>.sql` o `NNN_<accion>_<detalle>.sql` (ej. `007_make_prices_nullable.sql`).
- PK: `id UUID DEFAULT gen_random_uuid()` (excepto tablas de historial que usan `SERIAL`, ej. `order_status_history`).
- FK a usuario: `user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE`.
- Timestamps: `created_at`/`updated_at` TIMESTAMPTZ con default; trigger `update_<tabla>_updated_at` usando `update_updated_at_column()`.
- Índices: `idx_<tabla>_<columna>`.
- Multi-tenant: scope siempre por `user_id`, casi siempre con `UNIQUE(user_id, <campo>)`.

## Qué NO hacer
- NO uses una herramienta de migración externa (golang-migrate, goose): el proyecto corre SQL idempotente inline al startup.
- NO escribas migraciones destructivas sin `IF NOT EXISTS` / `IF EXISTS` — deben poder re-correr en cada boot.
- NO rompas la numeración secuencial ni reutilices un número existente.
- NO olvides el índice por `user_id` ni el trigger de `updated_at` en tablas nuevas.

## Comandos útiles
```bash
cd producer/vento_core
go run cmd/server/main.go     # aplica migraciones idempotentes al conectar
ls migrations/                # ver numeración existente antes de agregar una
cd ../../devops && docker compose up -d   # Postgres :5433 (vento/vento_secret, db vento_db)
```
