# База данных

В проекте используется реляционная база данных **PostgreSQL**. Для работы с ней применяется ORM **SQLAlchemy** (async) и система миграций **Alembic**.

## Схема данных (ER Diagram Description)

### Users (`users`)

Хранит информацию о пользователях.

| Поле              | Тип             | Описание                           |
| ----------------- | --------------- | ---------------------------------- |
| `id`              | Integer (PK)    | Уникальный идентификатор           |
| `email`           | String (Unique) | Email пользователя (логин)         |
| `hashed_password` | String          | Хеш пароля (bcrypt)                |
| `full_name`       | String          | Полное имя (опционально)           |
| `github_id`       | String          | ID пользователя GitHub (для OAuth) |
| `avatar_url`      | String          | Ссылка на аватар                   |
| `created_at`      | DateTime        | Дата регистрации                   |

### Projects (`projects`)

Проекты, созданные пользователями.

| Поле       | Тип          | Описание                            |
| ---------- | ------------ | ----------------------------------- |
| `id`       | Integer (PK) | Идентификатор проекта               |
| `user_id`  | Integer (FK) | Владелец проекта (`users.id`)       |
| `name`     | String       | Название проекта                    |
| `repo_url` | String       | Ссылка на репозиторий (опционально) |

### Analyses (`analyses`)

История запусков анализа кода.

| Поле         | Тип          | Описание                            |
| ------------ | ------------ | ----------------------------------- |
| `id`         | Integer (PK) | Идентификатор анализа               |
| `project_id` | Integer (FK) | Проект (`projects.id`)              |
| `status`     | String       | Статус (pending, completed, failed) |
| `created_at` | DateTime     | Время запуска                       |

### Vulnerabilities (`vulnerabilities`)

Найденные проблемы безопасности в рамках конкретного анализа.

| Поле          | Тип          | Описание                        |
| ------------- | ------------ | ------------------------------- |
| `id`          | Integer (PK) | Идентификатор уязвимости        |
| `analysis_id` | Integer (FK) | Анализ (`analyses.id`)          |
| `type`        | String       | Тип (например, "SQL Injection") |
| `severity`    | String       | Критичность (High, Medium, Low) |
| `file`        | String       | Файл с проблемой                |
| `line`        | String       | Номер строки или диапазон       |
| `description` | Text         | Подробное описание              |

## Миграции

Управление схемой БД происходит через Alembic.

- `alembic revision --autogenerate -m "message"`: Создать новую миграцию на основе изменений в моделях `sql_models.py`.
- `alembic upgrade head`: Применить все миграции к текущей БД.
