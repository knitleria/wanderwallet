package routes

import (
	"wanderwallet/internal/controllers"
	"wanderwallet/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	r *gin.Engine,
	userController *controllers.UserController,
	travelController *controllers.TravelController,
	expenseController *controllers.ExpenseController,
	categoryController *controllers.CategoryController,
	analyticsController *controllers.AnalyticsController,
) {

	api := r.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", userController.Register)
			authRoutes.POST("/login", userController.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware)
		{
			travelRoutes := protected.Group("/travel")
			{
				travelRoutes.POST("", travelController.CreateTravel)
			}

			expenseRoutes := protected.Group("/expenses")
			{
				expenseRoutes.GET("", expenseController.GetExpensesByUserID)
				expenseRoutes.POST("", expenseController.CreateExpense)
				expenseRoutes.PUT("/:id", expenseController.UpdateExpenseByUserID)
				expenseRoutes.DELETE("/:id", expenseController.DeleteExpenseByID)
			}

			categoryRoutes := protected.Group("/categories")
			{
				categoryRoutes.GET("", categoryController.GetCategoriesByUserID)
				categoryRoutes.POST("", categoryController.CreateCategory)
				categoryRoutes.DELETE("/:id", categoryController.DeleteCategoryByID)
			}

			analyticsRoutes := protected.Group("/analytics")
			{
				analyticsRoutes.GET("", analyticsController.GetAnalytics)
			}
		}

	}
}
