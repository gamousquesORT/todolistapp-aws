package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"todolistapp/internal/handlers"
	"todolistapp/internal/services"
)

func main() {
	// Get environment variables
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		tableName = "todos"
	}

	// Initialize DynamoDB service
	dbService, err := services.NewDynamoDBService(region, tableName)
	if err != nil {
		log.Fatal("Failed to initialize DynamoDB service:", err)
	}

	r := gin.Default()

	todoHandler := handlers.NewTodoHandler(dbService)

	api := r.Group("/api/v1")
	{
		api.GET("/todos", todoHandler.GetTodos)
		api.POST("/todos", todoHandler.CreateTodo)
		api.GET("/todos/:id", todoHandler.GetTodoByID)
		api.PUT("/todos/:id", todoHandler.UpdateTodo)
		api.DELETE("/todos/:id", todoHandler.DeleteTodo)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
