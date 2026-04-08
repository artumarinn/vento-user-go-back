package handler

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/infrastructure/adapter/postgres"
	"net/http"
)

type MetaHandler struct {
	metaRepo *postgres.MetaRepo
}

func NewMetaHandler(metaRepo *postgres.MetaRepo) *MetaHandler {
	return &MetaHandler{metaRepo: metaRepo}
}

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
