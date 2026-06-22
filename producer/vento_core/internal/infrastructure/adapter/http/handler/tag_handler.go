package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type TagHandler struct {
	tagUC *usecase.TagUsecases
}

func NewTagHandler(tagUC *usecase.TagUsecases) *TagHandler {
	return &TagHandler{tagUC: tagUC}
}

func (h *TagHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	tags, err := h.tagUC.ListTags(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

func (h *TagHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.tagUC.CreateTag(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tag)
}

func (h *TagHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	tagID := c.Param("id")

	if err := h.tagUC.DeleteTag(c.Request.Context(), userID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
