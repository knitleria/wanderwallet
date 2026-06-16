package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"wanderwallet/internal/dto"
	"wanderwallet/internal/models"
	"wanderwallet/internal/services"

	"github.com/gin-gonic/gin"
)

type TravelController struct {
	travelService *services.TravelService
}

func NewTravelController(travelService *services.TravelService) *TravelController {
	return &TravelController{
		travelService: travelService,
	}
}

// CreateTravel godoc
// @Summary Создать новое путешествие
// @Description Создает запись о путешествии для авторизованного пользователя
// @Tags travel
// @Accept json
// @Produce json
// @Param travel body dto.CreateTravelRequest true "Данные путешествия"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/travel [post]
func (ctrl *TravelController) CreateTravel(c *gin.Context) {
	var req dto.CreateTravelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}
	user := c.MustGet("user").(models.User)
	ctx := c.Request.Context()

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
		return
	}

	travel, err := ctrl.travelService.CreateTravel(ctx, user.ID, req.Title, startDate, endDate)
	if err != nil {
		log.Printf("Failed to create travel for user %d: %v\n", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Travel %s with %d created", travel.Title, travel.ID),
	})
}
