package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"wanderwallet/internal/dto"
	"wanderwallet/internal/events"
	"wanderwallet/internal/models"
	"wanderwallet/internal/services"

	"time"

	"github.com/gin-gonic/gin"
)

type ExpenseController struct {
	expenseService  *services.ExpenseService
	categoryService *services.CategoryService
	travelService   *services.TravelService
	eventPublisher  events.Publisher
}

func NewExpenseController(expenseService *services.ExpenseService, categoryService *services.CategoryService, travelService *services.TravelService, eventPublisher events.Publisher) *ExpenseController {
	return &ExpenseController{
		expenseService:  expenseService,
		categoryService: categoryService,
		travelService:   travelService,
		eventPublisher:  eventPublisher,
	}
}

// CreateExpense godoc
// @Summary Создать расход
// @Description Добавляет новый расход для авторизованного пользователя
// @Tags expenses
// @Accept json
// @Produce json
// @Param expense body dto.CreateExpenseRequest true "Данные расхода"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/expenses [post]
func (ctrl *ExpenseController) CreateExpense(c *gin.Context) {
	var req dto.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	user := c.MustGet("user").(models.User)
	ctx := c.Request.Context()

	category, err := ctrl.categoryService.GetCategoryByName(ctx, req.Category)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found or unavailable"})
		return
	}

	travel, err := ctrl.travelService.GetTravelByID(ctx, req.TravelID)
	if err != nil || travel.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid travel"})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}

	expense := &models.Expense{
		UserID:      user.ID,
		CategoryID:  category.ID,
		TravelID:    req.TravelID,
		Amount:      req.Amount,
		CreatedAt:   date,
		Description: req.Comment,
	}

	if err := ctrl.expenseService.CreateExpense(ctx, expense); err != nil {
		log.Printf("Failed to create expense for user %d: %v\n", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	if err := ctrl.eventPublisher.PublishExpense(ctx, events.ExpenseCreatedEvent{
		EventType: "expense_created",
		ExpenseId: expense.ID,
		UserID:    user.ID,
		TravelID:  expense.TravelID,
		Category:  category.Name,
		Amount:    expense.Amount,
		CreatedAt: expense.CreatedAt.Format(time.RFC3339),
	}); err != nil {
		log.Printf("Failed to publish expense created event: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Expense with amount %.2f created", expense.Amount),
	})
}

// GetExpensesByUserID godoc
// @Summary Получить расходы пользователя
// @Description Возвращает список расходов текущего пользователя по категории и дате (опционально)
// @Tags expenses
// @Accept json
// @Produce json
// @Param category query string false "Категория"
// @Param from query string false "Дата начала, формат YYYY-MM-DD"
// @Param to query string false "Дата окончания, формат YYYY-MM-DD"
// @Success 200 {array} dto.ExpenseResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/expenses [get]
func (ctrl *ExpenseController) GetExpensesByUserID(c *gin.Context) {
	var req dto.GetUsersExpenseRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query params"})
		return
	}
	user := c.MustGet("user").(models.User)
	ctx := c.Request.Context()

	var fromDate, toDate *time.Time
	if req.From != "" {
		t, err := time.Parse("2006-01-02", req.From)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date"})
			return
		}
		fromDate = &t
	}
	if req.To != "" {
		t, err := time.Parse("2006-01-02", req.To)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date"})
			return
		}
		toDate = &t
	}

	var categoryID *uint
	if req.Category != "" {
		cat, err := ctrl.categoryService.GetCategoryByName(ctx, req.Category)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		categoryID = &cat.ID
	}

	expenses, err := ctrl.expenseService.GetExpensesByUserTimeAndCategory(ctx, user.ID, fromDate, toDate, categoryID)
	if err != nil {
		log.Printf("Failed to get expenses for user %d: %v\n", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	expenseResponses := make([]dto.ExpenseResponse, 0, len(expenses))
	for _, e := range expenses {
		category, _ := ctrl.categoryService.GetCategoryByID(ctx, e.CategoryID)
		expenseResponses = append(expenseResponses, dto.ExpenseResponse{
			ID:       fmt.Sprintf("%v", e.ID),
			Category: category.Name,
			Amount:   e.Amount,
			Date:     e.CreatedAt.Format("2006-01-02"),
			Comment:  e.Description,
		})
	}
	c.JSON(http.StatusOK, expenseResponses)
}

// UpdateExpenseByUserID godoc
// @Summary Обновить расход
// @Description Обновляет данные расхода по его ID для текущего пользователя
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "ID расхода"
// @Param expense body dto.UpdateExpenseRequest true "Новые данные расхода"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/expenses/{id} [put]
func (ctrl *ExpenseController) UpdateExpenseByUserID(c *gin.Context) {
	idStr := c.Param("id")
	expenseID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense ID"})
		return
	}

	var req dto.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user := c.MustGet("user").(models.User)
	ctx := c.Request.Context()
	expense, err := ctrl.expenseService.GetExpenseByID(ctx, uint(expenseID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}

	if expense.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot edit another user's expense"})
		return
	}

	category, err := ctrl.categoryService.GetCategoryByName(ctx, req.Category)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	expenseDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}

	expense.CategoryID = category.ID
	expense.Amount = req.Amount
	expense.CreatedAt = expenseDate
	expense.Description = req.Comment

	if err := ctrl.expenseService.UpdateExpense(ctx, expense); err != nil {
		log.Printf("Failed to update expense %d: %v\n", expense.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense updated successfully",
	})
}

// DeleteExpenseByID godoc
// @Summary Удалить расход
// @Description Удаляет расход текущего пользователя по ID
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path int true "ID расхода"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/expenses/{id} [delete]
func (ctrl *ExpenseController) DeleteExpenseByID(c *gin.Context) {
	idStr := c.Param("id")
	expenseID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense ID"})
		return
	}

	user := c.MustGet("user").(models.User)
	ctx := c.Request.Context()

	expense, err := ctrl.expenseService.GetExpenseByID(ctx, uint(expenseID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}

	if expense.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete another user's expense"})
		return
	}

	if err := ctrl.expenseService.DeleteExpense(ctx, uint(expenseID)); err != nil {
		log.Printf("Failed to delete expense %d: %v\n", expense.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense deleted successfully",
	})
}
