package interfaces

import (
	"github.com/google/uuid"
	"homecloud-auth-service/internal/security"
)

type Security interface {
	// Хеширование паролей
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
	
	// Работа с токенами
	GenerateToken(userID uuid.UUID) (string, error)
	ValidateToken(tokenString string) (*security.TokenClaims, error)
	RefreshToken(tokenString string) (string, error)
	InvalidateToken(tokenString string) error
	
	// Генерация токенов верификации
	GenerateVerificationToken(userID uuid.UUID) (string, error)
	ValidateVerificationToken(tokenString string) (uuid.UUID, error)
} 