package usecase

import (
	"context"
	"errors"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

// RegisterUser orchestrates the user registration flow.
type RegisterUser struct {
	userRepo    port.UserRepository
	authService port.AuthService
}

// NewRegisterUser creates a RegisterUser use case with its dependencies.
func NewRegisterUser(repo port.UserRepository, auth port.AuthService) *RegisterUser {
	return &RegisterUser{
		userRepo:    repo,
		authService: auth,
	}
}

// Execute registers a new user, returning the auth response with a JWT token.
func (uc *RegisterUser) Execute(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user already exists
	existing, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Create domain entity (validates email, hashes password)
	user, err := entity.NewUser(req.Email, req.Password, req.FullName, req.BusinessName)
	if err != nil {
		return nil, err
	}

	// Persist
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, err
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
