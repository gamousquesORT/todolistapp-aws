package models

import (
	"time"
)

type Todo struct {
	ID          string    `json:"id" dynamodbav:"id"`
	Title       string    `json:"title" binding:"required" dynamodbav:"title"`
	Description string    `json:"description" dynamodbav:"description"`
	Completed   bool      `json:"completed" dynamodbav:"completed"`
	CreatedAt   time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type UpdateTodoRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}
