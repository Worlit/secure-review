#!/bin/bash

# Функция для завершения всех процессов при выходе (Ctrl+C)
cleanup() {
    echo -e "\n🛑 Останавливаем сервисы..."
    kill $(jobs -p) 2>/dev/null
    exit
}

trap cleanup SIGINT

echo "🚀 Запуск проекта Secure Review..."

# Запуск Backend
echo "🐍 Запускаем Backend (FastAPI)..."
(
    cd backend
    # Активация виртуального окружения, если оно есть
    if [ -f "venv/bin/activate" ]; then
        source venv/bin/activate
    fi
    # Запуск uvicorn с hot-reload
    uvicorn main:app --reload --port 8000
) &

# Запуск Frontend
echo "⚛️  Запускаем Frontend (Vite)..."
(
    cd frontend
    npm run dev
) &

# Ожидание завершения процессов
wait
