package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Security struct {
	jwtSecret string
	jwtExpiration time.Duration
	verificationSecret string
	verificationExpiration time.Duration
}

func NewSecurity(jwtSecret string, jwtExpiration time.Duration, verificationSecret string, verificationExpiration time.Duration) *Security {
	return &Security{
		jwtSecret: jwtSecret,
		jwtExpiration: jwtExpiration,
		verificationSecret: verificationSecret,
		verificationExpiration: verificationExpiration,
	}
}

// Хеширование паролей
func (s *Security) HashPassword(password string) (string, error) {
	fmt.Printf("DEBUG: Hashing password: %s\n", password)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	hashStr := string(hash)
	fmt.Printf("DEBUG: Generated hash: %s\n", hashStr)
	return hashStr, nil
}

func (s *Security) ComparePassword(hashedPassword, password string) error {
	fmt.Printf("DEBUG: Comparing password: %s with hash: %s\n", password, hashedPassword)
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		fmt.Printf("DEBUG: Password comparison failed: %v\n", err)
		return err
	}
	fmt.Printf("DEBUG: Password comparison successful\n")
	return nil
}

// Генерация случайного ID для токенов
func generateRandomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// JWT токены
func (s *Security) GenerateToken(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(s.jwtExpiration)
	
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"token_id": generateRandomID(),
		"exp": expirationTime.Unix(),
		"iat": time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	
	return tokenString, nil
}

func (s *Security) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	
	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, fmt.Errorf("token expired")
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id format: %w", err)
	}
	
	tokenID, _ := claims["token_id"].(string)
	
	return &TokenClaims{
		UserID: userID,
		TokenID: tokenID,
	}, nil
}

func (s *Security) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil && err.Error() != "token expired" {
		return "", err
	}
	
	return s.GenerateToken(claims.UserID)
}

func (s *Security) InvalidateToken(tokenString string) error {
	// В простой реализации просто валидируем токен
	// В реальном приложении можно хранить черный список токенов
	_, err := s.ValidateToken(tokenString)
	return err
}

// Токены верификации email
func (s *Security) GenerateVerificationToken(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(s.verificationExpiration)
	
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"type": "email_verification",
		"exp": expirationTime.Unix(),
		"iat": time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.verificationSecret))
	if err != nil {
		return "", fmt.Errorf("error signing verification token: %w", err)
	}
	
	return tokenString, nil
}

func (s *Security) ValidateVerificationToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.verificationSecret), nil
	})
	
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid verification token: %w", err)
	}
	
	if !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid verification token")
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid verification token claims")
	}
	
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "email_verification" {
		return uuid.Nil, fmt.Errorf("invalid token type")
	}
	
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user_id in verification token")
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user_id format: %w", err)
	}
	
	return userID, nil
}

type TokenClaims struct {
	UserID  uuid.UUID `json:"user_id"`
	TokenID string    `json:"token_id,omitempty"`
}