# Схема базы данных

## Таблицы

### users

Таблица пользователей системы.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255),
    github_id BIGINT UNIQUE,
    github_login VARCHAR(255),
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_github_id ON users(github_id);
```

| Поле          | Тип          | Описание                                  |
| ------------- | ------------ | ----------------------------------------- |
| id            | UUID         | Уникальный идентификатор                  |
| email         | VARCHAR(255) | Email пользователя (уникальный)           |
| username      | VARCHAR(50)  | Имя пользователя                          |
| password_hash | VARCHAR(255) | Хэш пароля (bcrypt), NULL для GitHub-only |
| github_id     | BIGINT       | ID пользователя на GitHub                 |
| github_login  | VARCHAR(255) | Логин на GitHub                           |
| avatar_url    | TEXT         | URL аватара                               |
| is_active     | BOOLEAN      | Активен ли аккаунт                        |
| created_at    | TIMESTAMP    | Дата создания                             |
| updated_at    | TIMESTAMP    | Дата обновления                           |

---

### code_reviews

Таблица запросов на анализ кода.

```sql
CREATE TABLE code_reviews (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    code TEXT NOT NULL,
    language VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    result TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_code_reviews_user_id ON code_reviews(user_id);
CREATE INDEX idx_code_reviews_status ON code_reviews(status);
```

| Поле         | Тип          | Описание                                     |
| ------------ | ------------ | -------------------------------------------- |
| id           | UUID         | Уникальный идентификатор                     |
| user_id      | UUID         | ID пользователя (FK)                         |
| title        | VARCHAR(255) | Название review                              |
| code         | TEXT         | Исходный код                                 |
| language     | VARCHAR(50)  | Язык программирования                        |
| status       | VARCHAR(20)  | Статус (pending/processing/completed/failed) |
| result       | TEXT         | JSON результат анализа                       |
| created_at   | TIMESTAMP    | Дата создания                                |
| updated_at   | TIMESTAMP    | Дата обновления                              |
| completed_at | TIMESTAMP    | Дата завершения анализа                      |

---

### security_issues

Таблица найденных уязвимостей.

```sql
CREATE TABLE security_issues (
    id UUID PRIMARY KEY,
    review_id UUID NOT NULL REFERENCES code_reviews(id) ON DELETE CASCADE,
    severity VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    line_start INTEGER,
    line_end INTEGER,
    suggestion TEXT NOT NULL,
    cwe VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_security_issues_review_id ON security_issues(review_id);
CREATE INDEX idx_security_issues_severity ON security_issues(severity);
```

| Поле        | Тип          | Описание                    |
| ----------- | ------------ | --------------------------- |
| id          | UUID         | Уникальный идентификатор    |
| review_id   | UUID         | ID review (FK)              |
| severity    | VARCHAR(20)  | Уровень серьёзности         |
| title       | VARCHAR(255) | Заголовок уязвимости        |
| description | TEXT         | Подробное описание          |
| line_start  | INTEGER      | Начальная строка            |
| line_end    | INTEGER      | Конечная строка             |
| suggestion  | TEXT         | Рекомендация по исправлению |
| cwe         | VARCHAR(50)  | CWE идентификатор           |
| created_at  | TIMESTAMP    | Дата создания               |

## Диаграмма связей

```
┌─────────────────┐
│     users       │
├─────────────────┤
│ id (PK)         │
│ email           │
│ username        │
│ password_hash   │
│ github_id       │
│ github_login    │
│ avatar_url      │
│ is_active       │
│ created_at      │
│ updated_at      │
└────────┬────────┘
         │
         │ 1:N
         │
         ▼
┌─────────────────┐
│  code_reviews   │
├─────────────────┤
│ id (PK)         │
│ user_id (FK)    │◄──────┐
│ title           │       │
│ code            │       │
│ language        │       │
│ status          │       │
│ result          │       │
│ created_at      │       │
│ updated_at      │       │
│ completed_at    │       │
└────────┬────────┘       │
         │                │
         │ 1:N            │
         │                │
         ▼                │
┌─────────────────┐       │
│ security_issues │       │
├─────────────────┤       │
│ id (PK)         │       │
│ review_id (FK)  │───────┘
│ severity        │
│ title           │
│ description     │
│ line_start      │
│ line_end        │
│ suggestion      │
│ cwe             │
│ created_at      │
└─────────────────┘
```

## Миграции

Миграции выполняются автоматически при запуске приложения через функцию `repository.RunMigrations()`.

При необходимости можно выполнить миграции вручную:

```bash
psql $DATABASE_URL -f migrations/001_initial.sql
```
