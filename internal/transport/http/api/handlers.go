package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"homecloud-auth-service/internal/interfaces"
	"homecloud-auth-service/internal/models"
)

type Handler struct {
	userService interfaces.UserService
}

func NewHandler(userService interfaces.UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

// Извлечение токена из заголовка Authorization
func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNotSupported
	}
	
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", http.ErrNotSupported
	}
	
	return authHeader[7:], nil
}

// Извлечение пользователя из контекста (после middleware аутентификации)
func getUserFromContext(r *http.Request) (*models.User, error) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		return nil, http.ErrNotSupported
	}
	return user, nil
}

// Регистрация нового пользователя
// POST /api/v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	user, _, err := h.userService.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := models.RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Аутентификация пользователя
// POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	user, token, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	
	response := models.LoginResponse{
		Token: token,
		User: &models.UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Role:     user.Role,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Получение профиля пользователя
// GET /api/v1/auth/me
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	response := models.ProfileResponse{
		ID:              user.ID,
		Email:           user.Email,
		Username:        user.Username,
		Role:            user.Role,
		IsActive:        user.IsActive,
		IsEmailVerified: user.IsEmailVerified,
		StorageQuota:    user.StorageQuota,
		UsedSpace:       user.UsedSpace,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Выход из системы
// POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token, err := extractToken(r)
	if err != nil {
		http.Error(w, "Authorization header required", http.StatusBadRequest)
		return
	}
	
	err = h.userService.Logout(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// Обновление профиля пользователя
// PATCH /api/v1/users/{id}
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Получение пользователя из контекста (аутентифицированный пользователь)
	currentUser, err := getUserFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Извлечение ID из URL
	vars := mux.Vars(r)
	userIDStr, ok := vars["id"]
	if !ok {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	
	// Проверка, что пользователь обновляет свой профиль
	if currentUser.ID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	
	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	err = h.userService.UpdateProfile(r.Context(), userID, req.Username, req.OldPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// Верификация email
// GET /api/v1/auth/verify?token=...
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Verification token is required", http.StatusBadRequest)
		return
	}
	
	err := h.userService.VerifyEmail(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// Middleware для аутентификации
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := extractToken(r)
		if err != nil {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		
		user, err := h.userService.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		// Добавляем пользователя в контекст
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
} 