package dto

import "time"

// RegisterRequest is the input DTO for user registration.
type RegisterRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	FullName     string `json:"full_name" binding:"required"`
	BusinessName string `json:"business_name" binding:"required"`
}

// LoginRequest is the input DTO for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents a safe (no password) user in API responses.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding: "required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding: "required"`
	Password string `json:"password" binding: "required,min=8"`
}

// AuthResponse is the output DTO for auth endpoints (register/login).
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
