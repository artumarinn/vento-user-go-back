package usecase

import (
	"context"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain"
)

// LoginUser orchestrates the user login flow.
type LoginUser struct {
	userRepo    port.UserRepository
	authService port.AuthService
}

// NewLoginUser creates a LoginUser use case with its dependencies.
func NewLoginUser(repo port.UserRepository, auth port.AuthService) *LoginUser {
	return &LoginUser{
		userRepo:    repo,
		authService: auth,
	}
}

// Execute authenticates a user by email/password and returns a JWT token.
func (uc *LoginUser) Execute(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by email
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Verify password
	if !user.Password.Verify(req.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate JWT
	token, err := uc.authService.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email.String(),
			FullName:  user.FullName,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}
