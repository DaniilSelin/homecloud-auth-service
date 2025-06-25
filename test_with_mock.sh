#!/bin/bash

# Тестирование с mock файловым сервисом
echo "Запуск тестирования с mock файловым сервисом"
echo "================================================"

# Устанавливаем переменную окружения для использования mock
export USE_MOCK_FILE_SERVICE=true

echo "Настройка:"
echo "  USE_MOCK_FILE_SERVICE=true"
echo ""

# Запускаем сервис в фоне
echo "Запуск auth service с mock файловым сервисом..."
go run cmd/server/main.go &
SERVER_PID=$!

# Ждем запуска сервиса
echo "Ожидание запуска сервиса..."
sleep 5

# Проверяем, что сервис запустился
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "Сервис не запустился"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi

echo "Сервис запущен успешно"
echo ""

# Запускаем тесты
echo "Запуск тестов регистрации..."
bash test_registration.sh

# Останавливаем сервис
echo ""
echo "Остановка сервиса..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "Тестирование завершено!" 