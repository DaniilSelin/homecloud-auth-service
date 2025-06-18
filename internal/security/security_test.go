package security

import (
	"VotingSystem/config"
	"fmt"
	"testing"
	"time"
)

func TestGetHashPswd(t *testing.T) {
	cfg := config.Config{}
	security := NewSecurity(cfg)

	testPassword := "testPassword123"
	hash, err := security.GetHashPswd(testPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == testPassword {
		t.Error("Hash should not be equal to the original password")
	}

	err = security.CompareHashAndPassword(hash, testPassword)
	if err != nil {
		t.Errorf("Failed to compare hash and password: %v", err)
	}

	err = security.CompareHashAndPassword(hash, "wrongPassword")
	if err == nil {
		t.Error("Expected error when comparing hash with incorrect password")
	}
}

func TestJWTTokens(t *testing.T) {
	cfg := config.Config{
		Jwt: config.JwtConfig{
			SecretKey:  "test-secret-key",
			Expiration: 15,
		},
	}
	security := NewSecurity(cfg)
	
	userID := "user123"
	
	token, err := security.GenerateToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Error("Generated token should not be empty")
	}
	
	fmt.Printf("Original token: %s\n", token)
	
	claims, err := security.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}
	
	_, err = security.ValidateToken("invalid.token.string")
	if err == nil {
		t.Error("Expected error when validating invalid token")
	}
	
	time.Sleep(5 * time.Millisecond)
	
	refreshedToken, err := security.RefreshToken(token)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}
	
	fmt.Printf("Refreshed token: %s\n", refreshedToken)
	
	if refreshedToken == "" {
		t.Error("Refreshed token should not be empty")
	}
	if refreshedToken == token {
		t.Error("Refreshed token should be different from the original token")
	}
}

func TestExpiredToken(t *testing.T) {
	cfg := config.Config{
		Jwt: config.JwtConfig{
			SecretKey:  "test-secret-key",
			Expiration: 1, 
		},
	}
	security := NewSecurity(cfg)
	
	userID := "user123"
	
	token, err := security.GenerateToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	
	time.Sleep(1 * time.Millisecond)
	
	refreshedToken, err := security.RefreshToken(token)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}
	if refreshedToken == "" {
		t.Error("Refreshed token should not be empty")
	}
	if refreshedToken == token {
		t.Error("Refreshed token should be different from the original token")
	}
}