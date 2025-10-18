package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest for user login (FLEXIBLE: email OR username)
type LoginRequest struct {
	Email    string `json:"email"`    // ✅ Optional
	Username string `json:"username"` // ✅ Optional
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest for updating user profile
type UpdateProfileRequest struct {
	Username string `json:"username,omitempty" binding:"omitempty,min=3,max=30"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
}

// ChangePasswordRequest for changing user password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// AuthResponse after successful login
type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// ErrorResponse for error messages
type ErrorResponse struct {
	Error string `json:"error"`
}
