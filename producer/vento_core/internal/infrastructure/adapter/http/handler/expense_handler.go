package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/usecase"
)

type ExpenseHandler struct {
	expenseUC *usecase.ExpenseUsecases
}

func NewExpenseHandler(expenseUC *usecase.ExpenseUsecases) *ExpenseHandler {
	return &ExpenseHandler{expenseUC: expenseUC}
}

func (h *ExpenseHandler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	expenses, err := h.expenseUC.ListExpenses(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, expenses)
}

func (h *ExpenseHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req dto.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := h.expenseUC.CreateExpense(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, expense)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	expenseID := c.Param("id")

	if err := h.expenseUC.DeleteExpense(c.Request.Context(), userID, expenseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
