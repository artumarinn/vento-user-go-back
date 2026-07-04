package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type LocationHandler struct {
	locationUC *usecase.LocationUsecases
}

func NewLocationHandler(locationUC *usecase.LocationUsecases) *LocationHandler {
	return &LocationHandler{locationUC: locationUC}
}

// locationErrorStatus maps known location errors to their HTTP status,
// falling back to 500 for anything unrecognized.
func locationErrorStatus(err error) int {
	switch {
	case errors.Is(err, usecase.ErrCannotDeleteDefaultLocation),
		errors.Is(err, usecase.ErrCannotDeleteLastLocation):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func (h *LocationHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	locations, err := h.locationUC.ListLocations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locations)
}

func (h *LocationHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	location, err := h.locationUC.CreateLocation(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, location)
}

func (h *LocationHandler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	locationID := c.Param("id")
	var req dto.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	location, err := h.locationUC.UpdateLocation(c.Request.Context(), userID, locationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, location)
}

func (h *LocationHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	locationID := c.Param("id")

	if err := h.locationUC.DeleteLocation(c.Request.Context(), userID, locationID); err != nil {
		c.JSON(locationErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
