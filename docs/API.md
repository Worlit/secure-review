# API Documentation

Полная документация REST API для Secure Review.

## Базовый URL

```
http://localhost:8080/api/v1
```

## Аутентификация

API использует JWT (JSON Web Tokens) для аутентификации. Токен должен передаваться в заголовке:

```
Authorization: Bearer <token>
```

---

## Endpoints

### Аутентификация

#### POST /auth/register

Регистрация нового пользователя.

**Request Body:**

```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "securepassword123"
}
```

**Response (201 Created):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "johndoe",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Errors:**

- `400 Bad Request` - Некорректные данные
- `409 Conflict` - Пользователь с таким email уже существует

---

#### POST /auth/login

Вход пользователя.

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (200 OK):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "johndoe",
    "github_login": "johndoe",
    "avatar_url": "https://avatars.githubusercontent.com/u/123",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Errors:**

- `400 Bad Request` - Некорректные данные
- `401 Unauthorized` - Неверный email или пароль

---

#### POST /auth/refresh

Обновление JWT токена. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

#### POST /auth/change-password

Смена пароля. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "old_password": "currentpassword",
  "new_password": "newpassword123"
}
```

**Response (200 OK):**

```json
{
  "message": "Password changed successfully"
}
```

**Errors:**

- `401 Unauthorized` - Неверный текущий пароль

---

### GitHub Integration

#### POST /auth/github/link

Связать текущий аккаунт с GitHub.

**Request Body:**

```json
{ "code": "github_oauth_code" }
```

#### DELETE /auth/github/link

Отвязать GitHub аккаунт.

#### GET /github/repos

Список репозиториев (через GitHub App или OAuth).

**Response:**

```json
[
  {
    "id": 123,
    "name": "repo-name",
    "full_name": "owner/repo-name",
    "html_url": "https://github.com/...",
    "private": true
  }
]
```

#### POST /github/webhook

Обработка событий от GitHub App. Подпись проверяется автоматически.

---

### Code Reviews

### GitHub OAuth

#### GET /auth/github

Получить URL для OAuth авторизации через GitHub.

**Headers (Optional):**

- `Authorization: Bearer <token>` - Если передан, инициирует процесс привязки GitHub аккаунта к текущему пользователю.

**Response (200 OK):**

```json
{
  "url": "https://github.com/login/oauth/authorize?client_id=...",
  "state": "random_state_string"
}
```

---

#### GET /auth/github/callback

Callback endpoint для GitHub OAuth (Plan B: Backend Redirect).
Используется, если GitHub настроен на редирект непосредственно на бэкенд.
Устанавливает сессионную cookie `access_token` и перенаправляет на главную страницу фронтенда.

**Query Parameters:**

- `code` - Код авторизации от GitHub
- `state` - State для защиты от CSRF

**Response (302 Found):**

Редирект на `FRONTEND_URL`.

---

#### POST /auth/github/callback

Callback endpoint для GitHub OAuth (Plan A: Frontend-first).
Используется, если GitHub редиректит на фронтенд, и фронтенд отправляет `code` на бэкенд.
Устанавливает сессионную cookie `access_token`.

**Request Body:**

```json
{
  "code": "github_auth_code",
  "state": "optional_state"
}
```

**Response (200 OK):**

```json
{
  "token": "eyJhbGci...",
  "user": { ... }
}
```

---

#### POST /auth/github/link

Привязать GitHub аккаунт к существующему пользователю. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "code": "github_auth_code"
}
```

**Response (200 OK):**

```json
{
  "message": "GitHub account linked successfully"
}
```

**Errors:**

- `409 Conflict` - GitHub аккаунт уже привязан к другому пользователю

---

#### DELETE /auth/github/link

Отвязать GitHub аккаунт от пользователя. Требует аутентификации и наличия пароля.

**Headers:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
{
  "message": "GitHub account unlinked successfully"
}
```

---

### GitHub Данные

#### GET /github/repos (или /users/repos)

Получение списка репозиториев пользователя с GitHub.
Endpoint `/users/repos` является алиасом для `/github/repos`.

**Headers:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
[
  {
    "id": 1296269,
    "name": "Hello-World",
    "full_name": "octocat/Hello-World",
    "html_url": "https://github.com/octocat/Hello-World",
    "description": "This your first repo!",
    "language": null,
    "private": false
  }
]
```

**Errors:**

- `401 Unauthorized` - Требуется авторизация
- `500 Internal Server Error` - Ошибка получения данных от GitHub

---

### Пользователи

Authorization: Bearer <token>

````

**Response (200 OK):**

```json
{
  "message": "GitHub account unlinked successfully"
}
````

**Errors:**

- `400 Bad Request` - У пользователя не установлен пароль

---

### Пользователи

#### GET /users/me

Получить профиль текущего пользователя. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "username": "johndoe",
  "github_login": "johndoe",
  "avatar_url": "https://avatars.githubusercontent.com/u/123",
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

#### PUT /users/me

Обновить профиль пользователя. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "username": "newusername",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

**Response (200 OK):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "username": "newusername",
  "avatar_url": "https://example.com/avatar.jpg",
  "is_active": true,
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

#### DELETE /users/me

Деактивировать аккаунт пользователя. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
{
  "message": "Account deactivated successfully"
}
```

---

### Code Review

#### POST /reviews

Создать новый code review. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Request Body:**

```json
{
  "title": "Check my Python function",
  "code": "def login(username, password):\n    query = f\"SELECT * FROM users WHERE username='{username}' AND password='{password}'\"\n    return db.execute(query)",
  "language": "python"
}
```

**Response (201 Created):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440001",
  "title": "Check my Python function",
  "code": "def login...",
  "language": "python",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Статусы:**

- `pending` - Ожидает анализа
- `processing` - Анализируется
- `completed` - Анализ завершён
- `failed` - Ошибка анализа

---

#### GET /reviews

Получить список code reviews текущего пользователя. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Query Parameters:**

- `page` - Номер страницы (default: 1)
- `page_size` - Размер страницы (default: 20, max: 100)

**Response (200 OK):**

```json
{
  "reviews": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "660e8400-e29b-41d4-a716-446655440001",
      "title": "Check my Python function",
      "code": "def login...",
      "language": "python",
      "status": "completed",
      "security_issues": [
        {
          "id": "770e8400-e29b-41d4-a716-446655440002",
          "review_id": "550e8400-e29b-41d4-a716-446655440000",
          "severity": "critical",
          "title": "SQL Injection",
          "description": "The code is vulnerable to SQL injection attacks...",
          "line_start": 2,
          "line_end": 2,
          "suggestion": "Use parameterized queries instead...",
          "cwe": "CWE-89",
          "created_at": "2024-01-15T10:31:00Z"
        }
      ],
      "created_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:31:00Z"
    }
  ],
  "total": 15,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

---

#### GET /reviews/:id

Получить детали code review. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440001",
  "title": "Check my Python function",
  "code": "def login(username, password):\n    query = f\"SELECT * FROM users WHERE username='{username}' AND password='{password}'\"\n    return db.execute(query)",
  "language": "python",
  "status": "completed",
  "result": "{\"summary\":\"Critical security vulnerability found...\",\"overall_score\":15}",
  "security_issues": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "review_id": "550e8400-e29b-41d4-a716-446655440000",
      "severity": "critical",
      "title": "SQL Injection Vulnerability",
      "description": "The code constructs SQL queries using string formatting with user input, making it vulnerable to SQL injection attacks.",
      "line_start": 2,
      "line_end": 2,
      "suggestion": "Use parameterized queries or an ORM to prevent SQL injection.",
      "cwe": "CWE-89",
      "created_at": "2024-01-15T10:31:00Z"
    }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:31:00Z"
}
```

**Errors:**

- `403 Forbidden` - Нет доступа к этому review
- `404 Not Found` - Review не найден

---

#### GET /reviews/:id/pdf

Скачивание отчета о проверке в формате PDF.

**Parameters:**

- `id` (path, required) - ID проверки

**Response (200 OK):**

- `Content-Type: application/pdf`
- `Content-Disposition: attachment; filename="review-<id>.pdf"`
- Binary PDF content

**Errors:**

- `404 Not Found` - Проверка не найдена
- `403 Forbidden` - Нет доступа к проверке
- `500 Internal Server Error` - Ошибка генерации PDF

---

#### DELETE /reviews/:id

Удалить code review. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
{
  "message": "Review deleted successfully"
}
```

---

#### POST /reviews/:id/reanalyze

Повторно проанализировать code review. Требует аутентификации.

**Headers:**

```
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440001",
  "title": "Check my Python function",
  "code": "def login...",
  "language": "python",
  "status": "pending",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### Health Check

#### GET /health

Проверка здоровья сервиса.

**Response (200 OK):**

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "database": "connected"
}
```

---

#### GET /ready

Проверка готовности сервиса.

**Response (200 OK):**

```json
{
  "message": "ready"
}
```

---

## Уровни серьёзности уязвимостей

| Уровень    | Описание                                                 |
| ---------- | -------------------------------------------------------- |
| `critical` | Критическая уязвимость, требует немедленного исправления |
| `high`     | Высокий риск, важно исправить                            |
| `medium`   | Средний риск                                             |
| `low`      | Низкий риск                                              |
| `info`     | Информационное сообщение                                 |

## Коды ошибок

| Код | Описание                                  |
| --- | ----------------------------------------- |
| 400 | Bad Request - Некорректные входные данные |
| 401 | Unauthorized - Требуется аутентификация   |
| 403 | Forbidden - Доступ запрещён               |
| 404 | Not Found - Ресурс не найден              |
| 409 | Conflict - Конфликт (например, дубликат)  |
| 500 | Internal Server Error - Внутренняя ошибка |

## Формат ошибок

```json
{
  "error": "Short error description",
  "message": "Detailed error message (optional)"
}
```
