package errdefs

import (
	"errors"
	"fmt"
)

var (
	//общие ошибки
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInternal     = errors.New("internal server error")

	// ошибки пакета security
	ErrGetHashPswd 	= errors.New("error hashing password")
	ErrGenerateToken = errors.New("failed to generate token")
	ErrByScript = errors.New("byscript error")

	// ошибки на уровне БД (repository)
	ErrDB			= errors.New("DB error")

	// ошибки слоя бизнесс логики (service)
	ErrExpiredToken = errors.New("token expired")
	ErrInvalidToken = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// Просто обертка, лучше в var добавить новую ошибку и использовать её
func New(text string) error {
	return errors.New(text)
}

// Wrap оборачивает ошибку с контекстом (аналог fmt.Errorf с %w)
func Wrap(err error, context string) error {
	return fmt.Errorf("%s: %w", context, err)
}

// Wrapf оборачивает ошибку с форматированием
func Wrapf(err error, format string, args ...interface{}) error {
    return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// Is проверяет соответствие ошибки (аналог errors.Is)
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As извлекает конкретный тип ошибки (аналог errors.As)
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}