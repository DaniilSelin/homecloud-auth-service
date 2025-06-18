# HomeCloud Auth Service

Сервис аутентификации для HomeCloud - микросервисная архитектура для управления пользователями и аутентификацией.

## Описание

Этот сервис предоставляет REST API для:
- Регистрации и аутентификации пользователей
- Управления профилями пользователей
- Верификации email
- Управления JWT токенами

## Архитектура

Сервис построен с использованием чистой архитектуры:

```
├── cmd/server/          # Точка входа в приложение
├── config/              # Конфигурация
├── internal/
│   ├── interfaces/      # Интерфейсы для всех слоев
│   ├── models/          # Модели данных
│   ├── repository/      # Слой доступа к данным (заглушка для gRPC)
│   ├── security/        # Безопасность и JWT
│   ├── service/         # Бизнес-логика
│   └── transport/http/  # HTTP API
└── README.md
```

## API Endpoints

### Аутентификация

| Метод | Путь | Описание | Вход / Выход |
|-------|------|----------|--------------|
| POST | `/api/v1/auth/register` | Регистрация нового пользователя | Request: `{ email, username, password }`<br>Response: `{ id, email, username, created_at }` |
| POST | `/api/v1/auth/login` | Аутентификация пользователя | Request: `{ email, password }`<br>Response: `{ token, user: { id, email, username, role } }` |
| GET | `/api/v1/auth/me` | Получить профиль пользователя | Response: `{ id, email, username, role, is_active, is_email_verified, storage_quota, used_space }` |
| POST | `/api/v1/auth/logout` | Выход из системы | — |
| GET | `/api/v1/auth/verify?token=...` | Верификация email | Response: 200 OK или 400 Bad Request |

### Управление профилем

| Метод | Путь | Описание | Вход / Выход |
|-------|------|----------|--------------|
| PATCH | `/api/v1/users/{id}` | Обновить профиль пользователя | Request: `{ username?, old_password?, new_password? }` |

## Модель пользователя

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Идентификация
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    
    -- Аутентификация и безопасность
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMP,
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,

    -- Информация о хранилище
    storage_quota BIGINT NOT NULL DEFAULT 10737418240, -- 10 GiB
    used_space BIGINT NOT NULL DEFAULT 0,

    -- Роли и разрешения
    role TEXT NOT NULL DEFAULT 'user', -- user / admin / readonly / etc
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,

    -- Метаданные
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);
```

## Установка и запуск

### Требования

- Go 1.22+
- Конфигурационный файл

### Конфигурация

Создайте файл `config/config.local.yaml` на основе `config/config.example.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

jwt:
  secret_key: "your-super-secret-jwt-key-change-in-production"
  expiration: "24h"

verification:
  secret_key: "your-super-secret-verification-key-change-in-production"
  expiration: "24h"

grpc:
  host: "localhost"
  port: 50051

logger:
  level: "info"
  encoding: "json"
  outputPaths: ["stdout"]
  errorOutputPaths: ["stderr"]
```

### Запуск

```bash
go mod tidy
go run cmd/server/main.go
```

## Особенности реализации

### Безопасность

- Пароли хешируются с использованием bcrypt
- JWT токены с настраиваемым временем жизни
- Защита от брутфорса (блокировка после 5 неудачных попыток)
- Отдельные токены для верификации email

### Архитектурные принципы

- **Интерфейсы**: Все зависимости определены через интерфейсы
- **Dependency Injection**: Сервисы получают зависимости через конструкторы
- **Clean Architecture**: Разделение на слои (transport, service, repository)
- **gRPC Ready**: Репозиторий подготовлен для работы через gRPC

### Временные заглушки

- Репозиторий содержит заглушки для gRPC вызовов
- В реальном приложении будет использоваться gRPC клиент для взаимодействия с сервисом БД

## Разработка

### Структура кода

- **interfaces/**: Определяет контракты между слоями
- **models/**: Модели данных и DTO
- **service/**: Бизнес-логика и валидация
- **repository/**: Доступ к данным (заглушка для gRPC)
- **security/**: JWT и хеширование паролей
- **transport/http/**: HTTP API и middleware

### Добавление новых функций

1. Определите интерфейс в `internal/interfaces/`
2. Реализуйте в соответствующем слое
3. Добавьте обработчик в `transport/http/api/`
4. Обновите маршруты в `routes.go`

## Лицензия

MIT
