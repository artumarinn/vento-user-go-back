package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type ClientHandler struct {
	clientUC *usecase.ClientUsecases
}

func NewClientHandler(clientUC *usecase.ClientUsecases) *ClientHandler {
	return &ClientHandler{clientUC: clientUC}
}

func (h *ClientHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	clients, err := h.clientUC.ListClients(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, clients)
}

func (h *ClientHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := h.clientUC.CreateClient(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, client)
}
