package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/valueobject"
)

// User is the core domain entity representing an authenticated user.
type User struct {
	ID           string
	Email        valueobject.Email
	Password     valueobject.Password // Optional for Google users
	FullName     string
	BusinessName string
	GoogleID     string // Optional
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser creates a new User with validated fields and a generated UUID.
func NewUser(email, plainPassword, fullName, businessName string) (*User, error) {
	if fullName == "" {
		return nil, domain.ErrEmptyFullName
	}

	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}

	passwordVO, err := valueobject.NewPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		ID:           uuid.New().String(),
		Email:        emailVO,
		Password:     passwordVO,
		FullName:     fullName,
		BusinessName: businessName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// NewGoogleUser creates a new User from Google profile information.
func NewGoogleUser(email, fullName, googleID string) (*User, error) {
	if fullName == "" {
		return nil, domain.ErrEmptyFullName
	}

	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		ID:        uuid.New().String(),
		Email:     emailVO,
		FullName:  fullName,
		GoogleID:  googleID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ReconstructUser rebuilds a User from persisted data.
func ReconstructUser(id, email, hashedPassword, fullName, businessName, googleID string, createdAt, updatedAt time.Time) *User {
	emailVO, _ := valueobject.NewEmail(email)
	return &User{
		ID:           id,
		Email:        emailVO,
		Password:     valueobject.NewPasswordFromHash(hashedPassword),
		FullName:     fullName,
		BusinessName: businessName,
		GoogleID:     googleID,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
