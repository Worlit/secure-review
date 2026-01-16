# Деплой на Railway

Проект подготовлен для деплоя через Docker контейнеры. Это самый надежный способ.

## Структура деплоя

Вам нужно создать **два** сервиса в Railway:

1.  **Backend (Database + API)**
2.  **Frontend (UI)**

---

## 1. Настройка Backend

1.  Создайте "New Service" -> "GitHub Repo" -> Выберите этот репозиторий.
2.  В настройках сервиса (Settings):
    *   **Root Directory**: `/backend`
    *   **Watch Paths**: `/backend/**` (опционально)
3.  **Variables** (Переменные окружения):
    *   Добавьте переменные из файла `.env`:
        *   `OPENAI_API_KEY`
        *   `SECRET_KEY`
        *   `PROJECT_NAME`
        *   `API_V1_STR`
        *   `GITHUB_CLIENT_ID` (если используете)
        *   `GITHUB_CLIENT_SECRET` (если используете)
    *   **DATABASE_URL**: Railway часто сам предоставляет PostgreSQL.
        *   Создайте "New Service" -> "Database" -> "PostgreSQL".
        *   Подключите его к сервису Backend, и переменная `DATABASE_URL` появится автоматически.
4.  Docker должен подхватиться автоматически благодаря наличию `Dockerfile` в папке `backend`.
    *   При старте контейнера автоматически применятся миграции (`alembic upgrade head`).

---

## 2. Настройка Frontend

1.  Создайте еще один "New Service" -> "GitHub Repo" -> Выберите *тот же самый* репозиторий.
2.  В настройках сервиса (Settings):
    *   **Root Directory**: `/frontend`
3.  **Variables**:
    *   Если фронтенду нужно знать URL бэкенда при сборке (Vite использует `VITE_` переменные), добавьте:
        *   `VITE_API_URL`: Ссылка на ваш задеплоенный Backend (например, `https://backend-production.up.railway.app`)
4.  Docker подхватится автоматически (`frontend/Dockerfile`).
    *   Проект соберется и будет раздаваться через Nginx.

---

## Важные нюансы

*   **Networking**: Frontend (в браузере пользователя) должен стучаться на **Публичный домен** бэкенда. Убедитесь, что вы сгенерировали домен для Backend сервиса в Railway (вкладка Settings -> Networking -> Generate Domain).
*   **CORS**: В `backend/main.py` в `allow_origins` нужно будет добавить домен вашего Frontend приложения после деплоя. Можно временно поставить `["*"]`, но лучше указать конкретный домен.

```python
# Пример обновления CORS в production
origins = [
    "http://localhost:5173",
    "https://your-frontend-app.up.railway.app" # <--- Добавьте свой домен
]
```
