# API Reference

Базовый URL: `/api/v1`

## Authentication

### Регистрация

`POST /auth/register`

Регистрация нового пользователя.

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "strongpassword",
  "full_name": "Ivan Ivanov"
}
```

**Response (200 OK):**

```json
{
  "id": 1,
  "email": "user@example.com",
  "full_name": "Ivan Ivanov",
  "github_id": null,
  "avatar_url": null
}
```

### Вход (Login)

`POST /auth/login/password`

Аутентификация по email и паролю.

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "strongpassword"
}
```

**Response (200 OK):**

```json
{
  "access_token": "token_string...",
  "token_type": "bearer"
}
```

## Analysis

### Анализ кода

`POST /analyze`

Отправка фрагмента кода на проверку безопасности.

**Request Body:**

```json
{
  "code": "import sqlite3\nconn = sqlite3.connect('example.db')\n...",
  "language": "python",
  "filename": "db.py"
}
```

**Response (200 OK):**

```json
{
  "issues": [
    {
      "type": "SQL Injection",
      "location": "Line 5",
      "description": "User input is directly concatenated into SQL query...",
      "severity": "High",
      "recommendation": "Use parameterized queries...",
      "fix_example": "cursor.execute('SELECT * FROM users WHERE id=?', (user_id,))"
    }
  ],
  "summary": "Code contains critical SQL injection vulnerability.",
  "security_score": 45
}
```
