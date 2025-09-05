package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"todolistapp/internal/models"
	"todolistapp/internal/services"
)

type TodoHandler struct {
	dbService *services.DynamoDBService
}

func NewTodoHandler(dbService *services.DynamoDBService) *TodoHandler {
	return &TodoHandler{
		dbService: dbService,
	}
}

func (h *TodoHandler) GetTodos(c *gin.Context) {
	todos, err := h.dbService.GetAllTodos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get todos: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  todos,
		"count": len(todos),
	})
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req models.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	todo := models.Todo{
		ID:          strconv.FormatInt(now.UnixNano(), 10),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := h.dbService.CreateTodo(&todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create todo: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": todo})
}

func (h *TodoHandler) GetTodoByID(c *gin.Context) {
	id := c.Param("id")
	
	todo, err := h.dbService.GetTodoByID(id)
	if err != nil {
		if err.Error() == "todo not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get todo: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": todo})
}

func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Completed != nil {
		updates["completed"] = *req.Completed
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	todo, err := h.dbService.UpdateTodo(id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update todo: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": todo})
}

func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id := c.Param("id")

	err := h.dbService.DeleteTodo(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete todo: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}