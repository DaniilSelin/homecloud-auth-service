package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"homecloud-auth-service/internal/interfaces"
	"homecloud-auth-service/internal/models"
	"homecloud-auth-service/internal/transport/grpc/fileClient"

	"github.com/google/uuid"
)

// Использованные ошибки -
// ErrDB - ошибка на уровне БД
// ErrByScript - Ошибка на уровне библиотеки bcrypt
// ErrNotFound - запись не найдена
// ErrGetHashPswd - ошибка ъеширования пароля
// ErrGenerateToken - ошибка генерации токена
// ErrInvalidToken - токен не раскодировался
// ErrExpiredToken - токен просрочен
// ErrInvalidCredentials - неправильные данные
// ErrEmailAlreadyExists - email уже существует

type UserService struct {
	repo        interfaces.UserRepository
	security    interfaces.Security
	fileService fileClient.FileServiceClient
}

func NewUserService(repo interfaces.UserRepository, security interfaces.Security, fileService fileClient.FileServiceClient) *UserService {
	return &UserService{
		repo:        repo,
		security:    security,
		fileService: fileService,
	}
}

// Регистрация нового пользователя
func (s *UserService) Register(ctx context.Context, email, username, password string) (*models.User, string, error) {
	fmt.Printf("DEBUG: Register called with email: %s, username: %s\n", email, username)

	// Валидация входных данных
	if err := s.validateRegistrationData(email, username, password); err != nil {
		fmt.Printf("DEBUG: Validation failed: %v\n", err)
		return nil, "", err
	}

	// Проверка существования email
	emailExists, err := s.repo.CheckEmailExists(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check email existence: %w", err)
	}
	if emailExists {
		return nil, "", fmt.Errorf("email already exists")
	}

	// Проверка существования username
	usernameExists, err := s.repo.CheckUsernameExists(ctx, username)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check username existence: %w", err)
	}
	if usernameExists {
		return nil, "", fmt.Errorf("username already exists")
	}

	// Хеширование пароля
	passwordHash, err := s.security.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Создание пользователя
	now := time.Now()
	user := &models.User{
		ID:              uuid.New(),
		Email:           email,
		Username:        username,
		PasswordHash:    passwordHash,
		IsActive:        true,
		IsEmailVerified: false,
		Role:            "user",
		IsAdmin:         false,
		StorageQuota:    10737418240, // 10 GiB
		UsedSpace:       0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Создание домашней директории для пользователя (обязательно)
	if s.fileService == nil {
		return nil, "", fmt.Errorf("file service is not available - cannot create user directory")
	}

	// Сначала создаем папку пользователя
	success, message, directoryPath, err := s.fileService.CreateUserDirectory(ctx, user.ID.String(), username)
	if err != nil {
		fmt.Printf("ERROR: Failed to create user directory: %v\n", err)
		return nil, "", fmt.Errorf("failed to create user directory: %w", err)
	}

	if !success {
		fmt.Printf("ERROR: File service returned failure: %s\n", message)
		return nil, "", fmt.Errorf("failed to create user directory: %s", message)
	}

	fmt.Printf("DEBUG: User directory created successfully: %s\n", directoryPath)

	// Теперь создаем пользователя в базе данных
	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		fmt.Printf("ERROR: Failed to create user in database after directory creation: %v\n", err)
		// TODO: Здесь можно добавить логику удаления созданной папки при неудаче создания пользователя
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Генерация JWT токена
	token, err := s.security.GenerateToken(userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	user.ID = userID
	fmt.Printf("DEBUG: User registered successfully: %s\n", user.Email)
	return user, token, nil
}

// Аутентификация пользователя
func (s *UserService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	fmt.Printf("DEBUG: Login called with email: %s\n", email)

	// Получение пользователя по email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		fmt.Printf("DEBUG: User not found by email: %s\n", email)
		return nil, "", fmt.Errorf("invalid credentials")
	}

	fmt.Printf("DEBUG: Found user: %s, password hash: %s\n", user.Email, user.PasswordHash)

	// Проверка активности и блокировки
	if !user.CanLogin() {
		fmt.Printf("DEBUG: User cannot login (locked or inactive)\n")
		return nil, "", fmt.Errorf("account is locked or inactive")
	}

	// Проверка пароля
	err = s.security.ComparePassword(user.PasswordHash, password)
	if err != nil {
		fmt.Printf("DEBUG: Password comparison failed: %v\n", err)
		// Увеличение счетчика неудачных попыток
		user.IncrementFailedAttempts()
		s.repo.UpdateFailedLoginAttempts(ctx, user.ID, user.FailedLoginAttempts)
		if user.LockedUntil != nil {
			s.repo.UpdateLockedUntil(ctx, user.ID, user.LockedUntil)
		}
		return nil, "", fmt.Errorf("invalid credentials")
	}

	fmt.Printf("DEBUG: Password comparison successful\n")

	// Сброс счетчика неудачных попыток
	if user.FailedLoginAttempts > 0 {
		user.ResetFailedAttempts()
		s.repo.UpdateFailedLoginAttempts(ctx, user.ID, 0)
		s.repo.UpdateLockedUntil(ctx, user.ID, nil)
	}

	// Обновление времени последнего входа
	now := time.Now()
	s.repo.UpdateLastLogin(ctx, user.ID)

	// Генерация JWT токена
	token, err := s.security.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	user.LastLoginAt = &now
	fmt.Printf("DEBUG: Login successful for user: %s\n", user.Email)
	return user, token, nil
}

// Валидация токена
func (s *UserService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	claims, err := s.security.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	return user, nil
}

// Выход из системы
func (s *UserService) Logout(ctx context.Context, token string) error {
	return s.security.InvalidateToken(token)
}

// Получение профиля пользователя
func (s *UserService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// Обновление профиля пользователя
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, username *string, oldPassword *string, newPassword *string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Обновление username
	if username != nil {
		if err := s.validateUsername(*username); err != nil {
			return err
		}

		usernameExists, err := s.repo.CheckUsernameExists(ctx, *username)
		if err != nil {
			return fmt.Errorf("failed to check username existence: %w", err)
		}
		if usernameExists && *username != user.Username {
			return fmt.Errorf("username already exists")
		}

		err = s.repo.UpdateUsername(ctx, userID, *username)
		if err != nil {
			return fmt.Errorf("failed to update username: %w", err)
		}
	}

	// Обновление пароля
	if newPassword != nil {
		if oldPassword == nil {
			return fmt.Errorf("old password is required to change password")
		}

		// Проверка старого пароля
		err = s.security.ComparePassword(user.PasswordHash, *oldPassword)
		if err != nil {
			return fmt.Errorf("invalid old password")
		}

		// Валидация нового пароля
		if err := s.validatePassword(*newPassword); err != nil {
			return err
		}

		// Хеширование нового пароля
		newPasswordHash, err := s.security.HashPassword(*newPassword)
		if err != nil {
			return fmt.Errorf("failed to hash new password: %w", err)
		}

		err = s.repo.UpdatePassword(ctx, userID, newPasswordHash)
		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}
	}

	return nil
}

// Верификация email
func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
	userID, err := s.security.ValidateVerificationToken(token)
	if err != nil {
		return fmt.Errorf("invalid verification token: %w", err)
	}

	err = s.repo.UpdateEmailVerification(ctx, userID, true)
	if err != nil {
		return fmt.Errorf("failed to update email verification: %w", err)
	}

	return nil
}

// Отправка email для верификации
func (s *UserService) SendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// В реальном приложении здесь была бы отправка email
	// Пока просто возвращаем успех
	return nil
}

// Обновление использования хранилища
func (s *UserService) UpdateStorageUsage(ctx context.Context, userID uuid.UUID, usedSpace int64) error {
	return s.repo.UpdateStorageUsage(ctx, userID, usedSpace)
}

// Получение пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// Валидация данных регистрации
func (s *UserService) validateRegistrationData(email, username, password string) error {
	if email == "" || username == "" || password == "" {
		return fmt.Errorf("all fields are required")
	}

	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email format")
	}

	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}

	return s.validatePassword(password)
}

// Валидация пароля
func (s *UserService) validatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	return nil
}

// Валидация username
func (s *UserService) validateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	return nil
}
