package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type MetaHandler struct {
	metaRepo port.MetaRepository
}

func NewMetaHandler(metaRepo port.MetaRepository) *MetaHandler {
	return &MetaHandler{metaRepo: metaRepo}
}

// saveMetaConfigRequest mirrors entity.MetaConfig fields that a user is allowed to set.
// user_id, id and timestamps come from the server, never the client.
type saveMetaConfigRequest struct {
	PlatformID           string  `json:"platform_id" binding:"required"`
	Channel              string  `json:"channel" binding:"required,oneof=whatsapp instagram messenger"`
	PermanentAccessToken string  `json:"permanent_access_token" binding:"required"`
	VerifyToken          *string `json:"verify_token"`
	AppSecret            *string `json:"app_secret"`
	WhatsAppBusinessID   *string `json:"whatsapp_business_id"`
}

// metaConfigResponse exposes the config without leaking secrets in full.
type metaConfigResponse struct {
	ID                   int64   `json:"id"`
	UserID               string  `json:"user_id"`
	PlatformID           string  `json:"platform_id"`
	Channel              string  `json:"channel"`
	WhatsAppBusinessID   *string `json:"whatsapp_business_id,omitempty"`
	PermanentAccessToken string  `json:"permanent_access_token"`
	VerifyToken          *string `json:"verify_token,omitempty"`
	AppSecret            *string `json:"app_secret,omitempty"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

func toResponse(c *entity.MetaConfig) metaConfigResponse {
	return metaConfigResponse{
		ID:                   c.ID,
		UserID:               c.UserID,
		PlatformID:           c.PlatformID,
		Channel:              c.Channel,
		WhatsAppBusinessID:   c.WhatsAppBusinessID,
		PermanentAccessToken: maskToken(c.PermanentAccessToken),
		VerifyToken:          maskOptional(c.VerifyToken),
		AppSecret:            maskOptional(c.AppSecret),
		CreatedAt:            c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:            c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func maskToken(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + strings.Repeat("•", 8) + s[len(s)-4:]
}

func maskOptional(s *string) *string {
	if s == nil {
		return nil
	}
	masked := maskToken(*s)
	return &masked
}

// Save persists or updates a Meta integration for the authenticated user.
// POST /api/v1/meta-configs
func (h *MetaHandler) Save(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	var req saveMetaConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Channel == "whatsapp" && (req.WhatsAppBusinessID == nil || strings.TrimSpace(*req.WhatsAppBusinessID) == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "whatsapp_business_id is required for whatsapp channel"})
		return
	}

	config := &entity.MetaConfig{
		UserID:               userID,
		PlatformID:           strings.TrimSpace(req.PlatformID),
		Channel:              req.Channel,
		PermanentAccessToken: strings.TrimSpace(req.PermanentAccessToken),
		VerifyToken:          req.VerifyToken,
		AppSecret:            req.AppSecret,
		WhatsAppBusinessID:   req.WhatsAppBusinessID,
	}

	if err := h.metaRepo.Save(c.Request.Context(), config); err != nil {
		log.Printf("[META] Save failed for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save meta config"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(config))
}

// ListMine returns all Meta configurations belonging to the authenticated user.
// GET /api/v1/meta-configs
func (h *MetaHandler) ListMine(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	configs, err := h.metaRepo.GetAllByUserID(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[META] ListMine failed for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list meta configs"})
		return
	}

	out := make([]metaConfigResponse, 0, len(configs))
	for _, cfg := range configs {
		out = append(out, toResponse(cfg))
	}
	c.JSON(http.StatusOK, out)
}

// GetByPlatformID is the internal lookup used by the IA service to resolve tokens.
// GET /api/v1/internal/meta-config/:platformID
func (h *MetaHandler) GetByPlatformID(c *gin.Context) {
	platformID := c.Param("platformID")
	if platformID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platformID is required"})
		return
	}

	config, err := h.metaRepo.GetByPlatformID(c.Request.Context(), platformID)
	if err != nil {
		log.Printf("Error fetching meta config for %s: %v", platformID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch meta config"})
		return
	}

	if config == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "meta config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}
