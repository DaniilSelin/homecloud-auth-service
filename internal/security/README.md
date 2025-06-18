# Пакет Security

Пакет security предоставляет функциональность для хэширования паролей и управления JWT-токенами для аутентификации и авторизации.

## Возможности

- Хэширование и проверка паролей с использованием bcrypt
- Генерация, валидация и обновление JWT-токенов
- Настраиваемое время жизни токенов

## Использование

### Создание экземпляра Security

```go
import (
    "VotingSystem/config"
    "VotingSystem/internal/security"
)

// Загрузка конфигурации приложения
cfg := config.Config{...}

// Создание нового экземпляра security
securityService := security.NewSecurity(cfg)
```

### Хэширование паролей

```go
// Хэширование пароля для хранения в базе данных
hashedPassword, err := securityService.GetHashPswd("user_password")
if err != nil {
    // Обработка ошибки
}

// Позже, проверка пароля на соответствие хэшу
err = securityService.CompareHashAndPassword(hashedPassword, "user_password")
if err != nil {
    // Пароль не совпадает
}
```

### Управление JWT-токенами

```go
// Генерация JWT-токена для пользователя
token, err := securityService.GenerateToken("user_id")
if err != nil {
    // Обработка ошибки
}

// Валидация токена
claims, err := securityService.ValidateToken(token)
if err != nil {
    if errors.Is(err, security.ErrExpiredToken) {
        // Обработка истёкшего токена
    } else if errors.Is(err, security.ErrInvalidToken) {
        // Обработка недействительного токена
    }
    // Обработка других ошибок
}

// Получение ID пользователя из claims
userID := claims.UserID

// Обновление токена (создание нового токена с продлённым сроком действия)
newToken, err := securityService.RefreshToken(token)
if err != nil {
    // Обработка ошибки
}
```

## Конфигурация

Пакет security использует следующую конфигурацию из структуры `config.Config`:

```go
type JwtConfig struct {
    SecretKey  string `yaml:"secret_key"`
    Expiration int    `yaml:"expiration"`
}
```

- `SecretKey`: Секретный ключ, используемый для подписи JWT-токенов
- `Expiration`: Время жизни токена в минутах

