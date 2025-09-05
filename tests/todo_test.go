package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"todolistapp/internal/handlers"
	"todolistapp/internal/models"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	
	todoHandler := handlers.NewTodoHandler()
	
	api := r.Group("/api/v1")
	{
		api.GET("/todos", todoHandler.GetTodos)
		api.POST("/todos", todoHandler.CreateTodo)
	}
	
	return r
}

func TestGetTodos(t *testing.T) {
	router := setupRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/todos", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "count")
}

func TestCreateTodo(t *testing.T) {
	router := setupRouter()
	
	todo := models.CreateTodoRequest{
		Title:       "Test Todo",
		Description: "This is a test todo item",
	}
	
	jsonData, _ := json.Marshal(todo)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 201, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
	
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Test Todo", data["title"])
	assert.Equal(t, "This is a test todo item", data["description"])
	assert.Equal(t, false, data["completed"])
	assert.NotNil(t, data["created_at"])
}

func TestCreateTodoWithoutTitle(t *testing.T) {
	router := setupRouter()
	
	todo := models.CreateTodoRequest{
		Description: "This todo has no title",
	}
	
	jsonData, _ := json.Marshal(todo)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/todos", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, 400, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}