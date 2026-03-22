package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain"
)

// AuthHandler exposes HTTP endpoints for authentication.
type AuthHandler struct {
	registerUC *usecase.RegisterUser
	loginUC    *usecase.LoginUser
	userRepo   port.UserRepository
}

// NewAuthHandler creates a new AuthHandler with the required use cases.
func NewAuthHandler(register *usecase.RegisterUser, login *usecase.LoginUser, userRepo port.UserRepository) *AuthHandler {
	return &AuthHandler{
		registerUC: register,
		loginUC:    login,
		userRepo:   userRepo,
	}
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	result, err := h.registerUC.Execute(c.Request.Context(), req)
	if err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	result, err := h.loginUC.Execute(c.Request.Context(), req)
	if err != nil {
		status := mapDomainError(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Me handles GET /api/v1/auth/me (requires authenticated user)
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email.String(),
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt,
	})
}

// mapDomainError translates domain errors to HTTP status codes.
func mapDomainError(err error) int {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrWeakPassword),
		errors.Is(err, domain.ErrEmptyFullName):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
