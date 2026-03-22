# Context: Vento Core Backend

## Purpose

The `vento-user-go-back` (internally named `vento_core`) is the primary backend service for the Vento.ai platform. It handles user authentication, product catalog management, and serves as the source of truth for business data in PostgreSQL.

## Tech Stack

| Component | Technology |
|-----------|-------------|
| **Language** | Go 1.25+ |
| **Web Framework** | Gin |
| **Database** | PostgreSQL (sqlx) |
| **Authentication** | JWT (HS256) |
| **Architecture** | Hexagonal (Clean Architecture) |

## Architecture

```
vento-user-go-back/producer/vento_core/
├── cmd/server/main.go                      # Entry point, wire dependencies
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── user.go                     # User aggregate root
│   │   │   └── product.go                  # Product entity
│   │   ├── valueobject/
│   │   │   ├── email.go                    # Email with validation
│   │   │   └── password.go                # Password with hashing
│   │   └── errors.go                       # Domain errors
│   ├── application/
│   │   ├── port/
│   │   │   ├── user_repository.go          # UserRepository interface
│   │   │   └── auth_service.go             # AuthService interface
│   │   ├── usecase/
│   │   │   ├── register_user.go            # Registration logic
│   │   │   ├── login_user.go               # Authentication logic
│   │   │   └── catalog_usecases.go         # Product CRUD operations
│   │   └── dto/
│   │       ├── auth.go                     # Auth request/response DTOs
│   │       └── product.go                  # Product DTOs
│   └── infrastructure/
│       ├── adapter/
│       │   ├── http/
│       │   │   ├── router.go               # Gin router setup
│       │   ├── http/handler/
│       │   │   ├── auth_handler.go         # Auth endpoints
│       │   │   └── product_handler.go      # Product endpoints
│       │   ├── http/middleware/
│       │   │   └── auth_middleware.go      # JWT validation
│       │   └── postgres/
│       │       ├── connection.go           # DB connection
│       │       ├── user_repo.go            # User persistence
│       │       └── product_repo.go         # Product persistence
│       ├── auth/
│       │   └── jwt.go                      # JWT service implementation
│       └── config/
│           └── config.go                   # Environment config
└── wiki/context.md                         # This file
```

## Core Responsibilities

### 1. User Authentication
- Registration with validated email and hashed password
- Login returning JWT token
- Token validation for protected routes

### 2. Product Catalog Management
- CRUD operations for products
- All products scoped to authenticated user (`user_id`)
- Stock tracking with units and supplier information

## HTTP Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| `GET` | `/health` | Health check | No |
| `POST` | `/api/v1/auth/register` | Create new account | No |
| `POST` | `/api/v1/auth/login` | Authenticate user | No |
| `GET` | `/api/v1/auth/me` | Get current user | Yes |
| `GET` | `/api/v1/products` | List user's products | Yes |
| `POST` | `/api/v1/products` | Create product | Yes |
| `PUT` | `/api/v1/products/:id` | Update product | Yes |
| `DELETE` | `/api/v1/products/:id` | Delete product | Yes |

## Domain Entities

### User
```go
type User struct {
    ID        string          // UUID
    Email     valueobject.Email    // Validated email
    Password  valueobject.Password // Hashed password
    FullName  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Factory methods
func NewUser(email, plainPassword, fullName string) (*User, error)  // Validates & hashes
func ReconstructUser(id, email, hashedPassword, fullName string, ...) *User  // From DB
```

### Product
```go
type Product struct {
    ID           string     // UUID
    UserID       string     // Owner reference
    Name         string     // Product name
    SKU          string     // Stock keeping unit
    Category     string     // Product category
    Stock        float64    // Current stock quantity
    StockUnit    string     // Unit: "u", "kg", "m", etc.
    MaxStock     *float64   // Optional max capacity
    Price        float64    // Selling price
    Supplier     string     // Supplier name
    SupplierCost float64    // Cost from supplier
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Value Objects

#### Email
```go
type Email struct {
    value string
}

func NewEmail(email string) (Email, error)  // Validates format
func (e Email) String() string
```

#### Password
```go
type Password struct {
    hash string
}

func NewPassword(plain string) (Password, error)  // Min 8 chars, hashes with bcrypt
func NewPasswordFromHash(hash string) Password
func (p Password) Hash() string
func (p Password) Matches(plain string) bool     // bcrypt comparison
```

## Ports (Interfaces)

### UserRepository
```go
type UserRepository interface {
    Save(ctx context.Context, user *entity.User) error
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    FindByID(ctx context.Context, id string) (*entity.User, error)
}
```

### AuthService
```go
type AuthService interface {
    GenerateToken(userID string) (string, error)
    ValidateToken(token string) (string, error)
    HashPassword(plain string) (string, error)
    ComparePassword(hashed, plain string) bool
}
```

## Database Schema

### users
```sql
CREATE TABLE users (
    id         VARCHAR(36) PRIMARY KEY,
    email      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255) NOT NULL,  -- bcrypt hash
    full_name  VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### products
```sql
CREATE TABLE products (
    id            VARCHAR(36) PRIMARY KEY,
    user_id       VARCHAR(36) REFERENCES users(id),
    name          VARCHAR(255) NOT NULL,
    sku           VARCHAR(100),
    category      VARCHAR(100),
    stock         DECIMAL(10, 2) DEFAULT 0,
    stock_unit    VARCHAR(20) DEFAULT 'u',
    max_stock     DECIMAL(10, 2),
    price         DECIMAL(10, 2) NOT NULL,
    supplier      VARCHAR(255),
    supplier_cost DECIMAL(10, 2),
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_user_id ON products(user_id);
```

## Configuration

Environment variables (loaded from `.env` or defaults):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5433` | PostgreSQL port |
| `DB_USER` | `vento` | Database user |
| `DB_PASSWORD` | `vento_secret` | Database password |
| `DB_NAME` | `vento_db` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_SECRET` | `vento-dev-secret...` | JWT signing secret |

**DSN Format:**
```
host=%s port=%s user=%s password=%s dbname=%s sslmode=%s
```

## JWT Token Format

```go
// Claims
type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}

// Token generation
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
signedToken, err := token.SignedString([]byte(secret))

// Expiration: 24 hours (configurable)
```

## Request/Response Examples

### Register
```json
// POST /api/v1/auth/register
{
  "email": "user@negocio.com",
  "password": "securepassword123",
  "full_name": "Juan Pérez"
}

// Response 201
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@negocio.com",
    "full_name": "Juan Pérez",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

### Login
```json
// POST /api/v1/auth/login
{
  "email": "user@negocio.com",
  "password": "securepassword123"
}

// Response 200
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { ... }
}
```

### Create Product
```json
// POST /api/v1/products
// Authorization: Bearer <token>
{
  "name": "Cartulina Opalina",
  "sku": "CART-OPAL-001",
  "category": "Papelería",
  "stock": 500,
  "stock_unit": "u",
  "price": 150.00,
  "supplier": "Distribuidora ABC",
  "supplier_cost": 80.00
}
```

## Dependencies

```go
// go.mod imports
github.com/gin-gonic/gin
github.com/jmoiron/sqlx
github.com/google/uuid
golang.org/x/crypto/bcrypt
github.com/golang-jwt/jwt/v5
```

## CORS Configuration

```go
// Allow all origins in development
r.Use(func(c *gin.Context) {
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
    c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
    if c.Request.Method == "OPTIONS" {
        c.AbortWithStatus(204)
        return
    }
    c.Next()
})
```

## Running Locally

```bash
# Prerequisites
- Go 1.25+
- PostgreSQL running on port 5433
- Database: vento_db

# Setup
cd vento-user-go-back/producer/vento_core
export DB_HOST=localhost
export DB_PORT=5433
export DB_USER=vento
export DB_PASSWORD=vento_secret
export DB_NAME=vento_db
export JWT_SECRET=your-secret-key

go run cmd/server/main.go

# Server starts on :8080
```

## Health Check Response

```json
{
  "status": "ok",
  "service": "vento_core"
}
```

## Integration Points

| Service | URL | Purpose |
|---------|-----|---------|
| Frontend | `http://localhost:3000` | Web UI |
| IA Service | `http://localhost:8081` | AI/ML operations |

## Known Limitations

1. **No Rate Limiting**: Endpoints are exposed without rate limiting
2. **No Request Validation**: DTOs lack comprehensive validation
3. **No Soft Delete**: Products are permanently deleted
4. **No Pagination**: Product list returns all items
5. **CORS Wildcard**: Allows all origins (configure for production)

## Future Improvements

- Add refresh token rotation
- Implement soft delete for products
- Add pagination and filtering to product list
- Add request validation middleware
- Implement rate limiting
- Add OpenAPI documentation