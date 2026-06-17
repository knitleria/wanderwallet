// @title Wanderwallet API
// @version 1.0
// @description API для управления расходами и путешествиями
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"wanderwallet/initializers"
	"wanderwallet/internal/config"
	"wanderwallet/internal/controllers"
	"wanderwallet/internal/middleware"
	"wanderwallet/internal/repository"
	"wanderwallet/internal/routes"
	"wanderwallet/internal/services"

	_ "wanderwallet/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {
	config.Init()
	cfg := config.Get()

	log.Println("Starting server on", cfg.RunAddress)

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "wanderwallet",
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Use(middleware.AuthMiddleware)
	r.Use(middleware.LoggerMiddleware())

	userRepo := repository.NewUserRepository(initializers.DB)
	travelRepo := repository.NewTravelRepository(initializers.DB)
	categoryRepo := repository.NewCategoryRepository(initializers.DB)
	expenseRepo := repository.NewExpenseRepository(initializers.DB)

	userService := services.NewUserService(userRepo)
	travelService := services.NewTravelService(travelRepo)
	categoryService := services.NewCategoryService(categoryRepo, expenseRepo)
	expenseService := services.NewExpenseService(expenseRepo)
	analyticsService := services.NewAnalyticsService(expenseRepo)

	userController := controllers.NewUserController(userService)
	travelController := controllers.NewTravelController(travelService)
	expenseController := controllers.NewExpenseController(expenseService, categoryService, travelService)
	categoryController := controllers.NewCategoryController(categoryService, expenseService)
	analyticsController := controllers.NewAnalyticsController(expenseService, analyticsService)

	routes.SetupRouter(r, userController, travelController, expenseController, categoryController, analyticsController)

	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server run failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting gracefully")
}
