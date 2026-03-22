package port

import (
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// AuthService defines the port for authentication token operations.
type AuthService interface {
	GenerateToken(user *entity.User) (string, error)
	ValidateToken(tokenString string) (userID string, err error)
}
