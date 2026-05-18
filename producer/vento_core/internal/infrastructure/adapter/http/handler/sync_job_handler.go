package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/entity"
)

type SyncJobHandler struct {
	repo port.SyncJobRepository
}

func NewSyncJobHandler(repo port.SyncJobRepository) *SyncJobHandler {
	return &SyncJobHandler{repo: repo}
}

type CreateJobRequest struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
}

func (h *SyncJobHandler) CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job := &entity.SyncJob{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Type:      req.Type,
		Status:    entity.SyncJobStatusQueued,
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.repo.Create(c.Request.Context(), job); err != nil {
		logger.L().Error("Failed to create sync job", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
		return
	}

	c.JSON(http.StatusCreated, job)
}

type UpdateProgressRequest struct {
	Progress int     `json:"progress"`
	Status   string  `json:"status"`
	ErrorMsg *string `json:"error_msg"`
}

func (h *SyncJobHandler) UpdateProgress(c *gin.Context) {
	jobID := c.Param("id")

	var req UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateProgress(c.Request.Context(), jobID, req.Progress, req.Status, req.ErrorMsg); err != nil {
		logger.L().Error("Failed to update sync job", "job_id", jobID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update job"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *SyncJobHandler) GetJob(c *gin.Context) {
	jobID := c.Param("id")

	job, err := h.repo.GetByID(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}
