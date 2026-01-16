# Установка и запуск

## Предварительные требования

Для работы проекта необходимо установить:

- **Python 3.10+** (для бэкенда)
- **Node.js 18+** (для фронтенда)
- **PostgreSQL** (можно использовать локальный сервер или облачный сервис, например, Railway/Supabase)

## Настройка окружения

1.  **Клонируйте репозиторий:**

    ```bash
    git clone https://github.com/your-username/secure-review.git
    cd secure-review
    ```

2.  **Создайте файл `.env` в корне проекта:**
    Вы можете скопировать пример (если он есть) или создать новый файл.

    ```ini
    # .env
    PROJECT_NAME="Secure Review"
    API_V1_STR="/api/v1"

    # Секретный ключ для сессий/токенов (сгенерируйте случайную строку)
    SECRET_KEY="your-super-secret-key"

    # OpenAI API Key (обязательно для анализа)
    OPENAI_API_KEY="sk-..."

    # Подключение к базе данных PostgreSQL
    # Формат: postgresql+asyncpg://user:password@host:port/dbname
    DATABASE_URL="postgresql+asyncpg://postgres:password@localhost:5432/securereview"

    # GitHub OAuth (опционально)
    GITHUB_CLIENT_ID=""
    GITHUB_CLIENT_SECRET=""
    ```

## Настройка Backend

1.  Перейдите в папку `backend`:

    ```bash
    cd backend
    ```

2.  Создайте и активируйте виртуальное окружение:

    ```bash
    python3 -m venv venv
    source venv/bin/activate  # macOS/Linux
    # venv\Scripts\activate   # Windows
    ```

3.  Установите зависимости:

    ```bash
    pip install -r requirements.txt
    ```

4.  Примените миграции базы данных:
    ```bash
    alembic upgrade head
    ```

## Настройка Frontend

1.  Перейдите в папку `frontend`:

    ```bash
    cd frontend
    ```

2.  Установите зависимости:
    ```bash
    npm install
    # или
    yarn install
    ```

## Запуск проекта

### Вариант 1: Автоматический (рекомендуемый)

Используйте скрипт `start.sh` из корня проекта, который запустит backend и frontend параллельно:

```bash
cd .. # Вернуться в корень проекта
chmod +x start.sh # Только при первом запуске
./start.sh
```

### Вариант 2: Ручной

**Backend:**

```bash
cd backend
source venv/bin/activate
uvicorn main:app --reload
```

API будет доступно по адресу: `http://localhost:8000`
Документация Swagger: `http://localhost:8000/docs`

**Frontend:**

```bash
cd frontend
npm run dev
```

Приложение откроется по адресу: `http://localhost:5173` (или другой порт, указанный в консоли)
