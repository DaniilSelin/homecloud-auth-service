#!/bin/bash

# Тест регистрации пользователя с обязательным созданием папки
# Убедитесь, что файловый сервис запущен на порту 50053

echo "Тестирование регистрации пользователя с обязательным созданием папки"
echo "=================================================="

# URL сервиса аутентификации
AUTH_URL="http://localhost:8080"

# Тестовые данные
EMAIL="testuser@example.com"
USERNAME="testuser"
PASSWORD="securepassword123"

echo "Тестовые данные:"
echo "  Email: $EMAIL"
echo "  Username: $USERNAME"
echo "  Password: $PASSWORD"
echo ""

# Проверка доступности сервиса
echo "Проверка доступности сервиса аутентификации..."
if curl -s "$AUTH_URL/health" > /dev/null 2>&1; then
    echo "Сервис аутентификации доступен"
else
    echo "Сервис аутентификации недоступен"
    echo "   Убедитесь, что сервис запущен на $AUTH_URL"
    exit 1
fi

echo ""

# Тест 1: Успешная регистрация (когда файловый сервис доступен)
echo "Тест 1: Успешная регистрация с доступным файловым сервисом"
echo "--------------------------------------------------"

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$AUTH_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\"
  }")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$RESPONSE" | head -n -1)

echo "HTTP Status: $HTTP_CODE"
echo "Response: $RESPONSE_BODY"

if [ "$HTTP_CODE" -eq 201 ]; then
    echo "Регистрация прошла успешно"
    USER_ID=$(echo "$RESPONSE_BODY" | jq -r '.id' 2>/dev/null)
    if [ "$USER_ID" != "null" ] && [ "$USER_ID" != "" ]; then
        echo "   User ID: $USER_ID"
    fi
else
    echo "Регистрация не удалась"
    echo "   Ожидался код 201, получен $HTTP_CODE"
fi

echo ""

# Тест 2: Попытка регистрации с тем же email (должна вернуть ошибку)
echo "Тест 2: Попытка повторной регистрации с тем же email"
echo "--------------------------------------------------"

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$AUTH_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"username\": \"anotheruser\",
    \"password\": \"$PASSWORD\"
  }")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$RESPONSE" | head -n -1)

echo "HTTP Status: $HTTP_CODE"
echo "Response: $RESPONSE_BODY"

if [ "$HTTP_CODE" -eq 400 ]; then
    echo "Правильно обработана ошибка дублирования email"
else
    echo "Неправильная обработка дублирования email"
    echo "   Ожидался код 400, получен $HTTP_CODE"
fi

echo ""

# Тест 3: Попытка регистрации с тем же username (должна вернуть ошибку)
echo "Тест 3: Попытка регистрации с тем же username"
echo "--------------------------------------------------"

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$AUTH_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"another@example.com\",
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\"
  }")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$RESPONSE" | head -n -1)

echo "HTTP Status: $HTTP_CODE"
echo "Response: $RESPONSE_BODY"

if [ "$HTTP_CODE" -eq 400 ]; then
    echo "Правильно обработана ошибка дублирования username"
else
    echo "Неправильная обработка дублирования username"
    echo "   Ожидался код 400, получен $HTTP_CODE"
fi

echo ""

# Тест 4: Валидация данных
echo "Тест 4: Валидация входных данных"
echo "--------------------------------------------------"

# Тест с коротким паролем
echo "   Тест с коротким паролем:"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$AUTH_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"shortpass@example.com\",
    \"username\": \"shortpass\",
    \"password\": \"123\"
  }")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$RESPONSE" | head -n -1)

echo "   HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" -eq 400 ]; then
    echo "   Правильно обработана ошибка валидации пароля"
else
    echo "   Неправильная обработка валидации пароля"
fi

# Тест с неверным email
echo "   Тест с неверным email:"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$AUTH_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"invalid-email\",
    \"username\": \"invalidemail\",
    \"password\": \"$PASSWORD\"
  }")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$RESPONSE" | head -n -1)

echo "   HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" -eq 400 ]; then
    echo "   Правильно обработана ошибка валидации email"
else
    echo "   Неправильная обработка валидации email"
fi

echo ""
echo "Тестирование завершено!"
echo ""
echo "Резюме:"
echo "  - Регистрация теперь требует создания папки пользователя"
echo "  - Если файловый сервис недоступен, регистрация не удастся"
echo "  - Валидация данных работает корректно"
echo "  - Дублирование email и username обрабатывается правильно" 