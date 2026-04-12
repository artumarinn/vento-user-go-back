package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
	"net/http"
)

type BusinessHandler struct {
	usecases *usecase.BusinessProfileUsecases
}

func NewBusinessHandler(usecases *usecase.BusinessProfileUsecases) *BusinessHandler {
	return &BusinessHandler{usecases: usecases}
}

func (h *BusinessHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	profile, err := h.usecases.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get business profile"})
		return
	}

	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "business profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *BusinessHandler) GetProfileInternal(c *gin.Context) {
	userID := c.Param("userID")
	profile, err := h.usecases.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get business profile"})
		return
	}

	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "business profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *BusinessHandler) SaveProfile(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var profile entity.BusinessProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	profile.UserID = userID
	if err := h.usecases.SaveProfile(c.Request.Context(), &profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save business profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *BusinessHandler) GetStats(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	stats, err := h.usecases.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get business stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
