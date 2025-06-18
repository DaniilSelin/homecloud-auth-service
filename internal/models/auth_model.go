package models

import (
	"github.com/google/uuid"
	"time"
)

// Запросы
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	Username    *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	OldPassword *string `json:"old_password,omitempty"`
	NewPassword *string `json:"new_password,omitempty" validate:"omitempty,min=6"`
}

// Ответы
type RegisterResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  *UserInfo  `json:"user"`
}

type UserInfo struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Username string `json:"username"`
	Role   string   `json:"role"`
}

type ProfileResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	Username        string     `json:"username"`
	Role            string     `json:"role"`
	IsActive        bool       `json:"is_active"`
	IsEmailVerified bool       `json:"is_email_verified"`
	StorageQuota    int64      `json:"storage_quota"`
	UsedSpace       int64      `json:"used_space"`
}

// JWT Claims
type Claims struct {
	UserID  uuid.UUID `json:"user_id"`
	TokenID string    `json:"token_id,omitempty"`
	Exp     int64     `json:"exp"`
	Iat     int64     `json:"iat"`
}