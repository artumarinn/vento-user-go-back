---
name: api-patterns
description: Estructura de handlers Gin, middleware de auth (JWT e internal-token), forma de respuestas de error y validación de DTOs en vento-user-go-back. Usar al agregar o modificar endpoints HTTP del Core.
---

## Cuándo usar esta skill
- Agregás o modificás un handler HTTP en el Core.
- Tocás middleware de autenticación o registro de rutas en `router.go`.
- Creás un nuevo usecase con inyección de dependencias vía port.

## Patrones establecidos

**Handler típico** (`internal/infrastructure/adapter/http/handler/product_handler.go`): leer `userID` del contexto, bind+validar DTO, llamar usecase, responder.
```go
func (h *ProductHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	product, err := h.catalogUC.CreateProduct(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, product)
}
```

**Error de cliente**: `gin.H{"error": err.Error()}` (handlers con JWT). Las rutas internas usan forma anidada: `gin.H{"error": gin.H{"code": "...", "message": "..."}}`.

**Middleware JWT** (`middleware/auth_middleware.go`): valida `Authorization: Bearer`, extrae userID y `c.Set("userID", userID)`. Aborta con `c.AbortWithStatusJSON(401, ...)`.

**Middleware internal** (`middleware/internal_token.go`): compara `X-Internal-Token` con `subtle.ConstantTimeCompare`, exige `X-Tenant-User-Id`, setea `c.Set("tenantUserID", tenantID)`.

**Registro de rutas** (`router.go`) — grupos con middleware atado:
```go
v1 := r.Group("/api/v1")
internal := v1.Group("/internal")
internal.Use(middleware.InternalTokenMiddleware(internalToken))
internal.POST("/products/sync", productHandler.SyncFromIA)

products := v1.Group("/products")
products.Use(middleware.AuthMiddleware(authService))
products.POST("", productHandler.Create)
```

**Usecase con DI por port** (`internal/application/usecase/catalog_usecases.go`):
```go
type CatalogUsecases struct {
	repo    port.ProductRepository
	tagRepo port.TagRepository
}
func NewCatalogUsecases(repo port.ProductRepository, tagRepo port.TagRepository) *CatalogUsecases {
	return &CatalogUsecases{repo: repo, tagRepo: tagRepo}
}
```

## Convenciones de nombrado
- Handlers: `<Recurso>Handler`, métodos `Create/List/Update/Delete` + `SyncFromIA`/`ListInternal` para rutas internas.
- Usecases: `<Dominio>Usecases` con constructor `New<Dominio>Usecases`.
- DTOs en `application/dto`, tags `json` + `binding` para validación.
- Repos: interfaz `port.<X>Repository`, implementación `Postgres<X>Repository` en `infrastructure/adapter/postgres`.
- Contexto: `userID` (JWT) vs `tenantUserID` (internal). No mezclar.

## Qué NO hacer
- NO instanciar repos concretos dentro de un usecase: inyectá la interfaz `port.X` por constructor.
- NO leer `userID` con `c.GetString` en rutas internas: ahí el dato vive en `tenantUserID`.
- NO devolver structs de entidad crudos en endpoints con JWT: mapeá a DTO (`mapEntityToDTO`).
- NO atar un handler internal al grupo con JWT ni viceversa.
- NO usar binding tags con espacios (`binding: "required, email"`) — rompe la validación silenciosamente (bug conocido en ForgotPassword/ResetPassword).

## Comandos útiles
```bash
cd producer/vento_core
go build ./...
go run cmd/server/main.go                                  # :8082 (env SERVER_PORT)
go test ./...
go test ./internal/domain/entity -run TestNewProduct -v    # test único
cd ../../devops && docker compose up -d                     # Postgres :5433
```
