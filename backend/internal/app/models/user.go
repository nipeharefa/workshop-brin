package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Phone     string    `json:"phone" db:"phone"`
	Email     string    `json:"email" db:"email"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Phone string `json:"phone" binding:"required,min=10,max=20"`
	Email string `json:"email" binding:"required,email,max=100"`
}

type UpdateUserRequest struct {
	Name     string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Email    string `json:"email,omitempty" binding:"omitempty,email,max=100"`
	IsActive *bool  `json:"is_active,omitempty"`
}