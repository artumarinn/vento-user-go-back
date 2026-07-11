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

func (h *BusinessHandler) UpdateAgentMode(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var body struct {
		DefaultAgentMode string `json:"default_agent_mode"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if body.DefaultAgentMode != "autonomous" && body.DefaultAgentMode != "assisted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "default_agent_mode must be 'autonomous' or 'assisted'"})
		return
	}
	if err := h.usecases.UpdateAgentMode(c.Request.Context(), userID, body.DefaultAgentMode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update agent mode"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"default_agent_mode": body.DefaultAgentMode})
}

func (h *BusinessHandler) UpdateTone(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var body struct {
		Tone string `json:"tone"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	validTones := map[string]bool{
		"rioplatense": true, "amigable": true, "profesional": true, "tecnico": true,
	}
	if !validTones[body.Tone] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tone must be one of: rioplatense, amigable, profesional, tecnico"})
		return
	}
	if err := h.usecases.UpdateTone(c.Request.Context(), userID, body.Tone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tone"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tone": body.Tone})
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
