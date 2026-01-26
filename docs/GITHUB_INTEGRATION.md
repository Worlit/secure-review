# Интеграция с GitHub

Проект поддерживает два способа интеграции с GitHub:

1. **GitHub OAuth** — для аутентификации пользователей.
2. **GitHub App** — для доступа к репозиториям, получения Webhook-событий и автоматического анализа кода.

## 1. Создание GitHub App

GitHub App является предпочтительным способом интеграции, так как позволяет получать более детальные права доступа и работать с вебхуками без создания "технического пользователя".

### Шаги создания:

1. Перейдите в [GitHub Developer Settings](https://github.com/settings/apps).
2. Нажмите **"New GitHub App"**.
3. Заполните основные поля:
   - **GitHub App name**: `Secure Review`
   - **Homepage URL**: `http://localhost:5173`
   - **Callback URL**:
     - Вариант A (Frontend): `http://localhost:5173/auth/github/callback` (Рекомендуется)
     - Вариант B (Backend): `http://localhost:8080/api/v1/auth/github/callback`
   - **Webhook URL**: `http://localhost:8080/api/v1/github/webhook`
   - **Webhook Secret**: Генерируйте случайную строку (запишите её, она понадобится в `.env`).

4. **Permissions (Права доступа):**
   - **Repository permissions:**
     - `Contents`: Read-only (для чтения кода)
     - `Pull requests`: Read-only (для анализа PR)
     - `Metadata`: Read-only (обязательно)
   - **User permissions:**
     - `Email addresses`: Read-only (для сопоставления пользователей)

5. **Subscribe to events (События):**
   - `Pull request`
   - `Push`

6. Нажмите **"Create GitHub App"**.

### Получение ключей:

1. На странице настроек созданного приложение:
   - Скопируйте **App ID**.
   - Скопируйте **Client ID**.
   - Сгенерируйте и скопируйте **Client Secret**.
   - В разделе **Private keys** нажмите **Generate a private key**. Скачается `.pem` файл. Откройте его и скопируйте содержимое полностью.

## 2. Настройка переменных окружения

Добавьте полученные данные в `.env` файл:

```env
# GitHub App Configuration
GITHUB_APP_ID=123456
GITHUB_APP_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
GITHUB_WEBHOOK_SECRET=your_webhook_secret

# GitHub OAuth Configuration (используется Client ID/Secret от того же GitHub App)
GITHUB_CLIENT_ID=Iv1.xxxxxxxxxxx
GITHUB_CLIENT_SECRET=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
GITHUB_REDIRECT_URL=http://localhost:5173/auth/github/callback
```

> **Примечание:** `GITHUB_REDIRECT_URL` должен совпадать с тем, что указан в настройках GitHub App (Callback URL).

- Если используете Frontend-first подход (рекомендуемый), укажите URL фронтенда.
- Если используете прямой редирект на бэкенд, укажите URL бэкенда (`.../api/v1/auth/github/callback`).

## 3. Установка приложения

Чтобы приложение могло анализировать репозитории:

1. Перейдите на страницу публичного профиля вашего GitHub App (`https://github.com/apps/your-app-name`).
2. Нажмите **"Install"**.
3. Выберите аккаунт или организацию и репозитории, к которым нужно дать доступ.

После установки GitHub отправит вебхук `installation`, и информация о установке сохранится в базе данных Secure Review.

## 4. Схема работы

1. **Аутентификация:** Пользователь входит через GitHub OAuth. Мы проверяем `github_id` пользователя.
2. **Доступ к репозиториям:**
   - Сервис проверяет, установлено ли GitHub App для данного пользователя (по `github_id` или `account_login`).
   - Если установлено, используется токен установки (Installation Token) — это позволяет работать с API лимитами приложения, а не пользователя.
   - Если не установлено, используется OAuth токен пользователя (fallback).

## 5. Webhooks

Эндпоинт: `POST /api/v1/github/webhook`

Обрабатываемые события:

- `installation`: Создание/удаление установки приложения.
- `installation_repositories`: Добавление/удаление репозиториев из установки.
- `push`: (В планах) Авто-анализ при пуше.
- `pull_request`: (В планах) Авто-анализ PR.
