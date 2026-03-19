# Схема базы данных

## ORM

Проект использует **GORM** — Go ORM с TypeORM-подобным API. Миграции выполняются автоматически через `AutoMigrate()`.

## Entity

### User Entity

```go
// internal/entity/user.go
type User struct {
    ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Email        string         `gorm:"size:255;uniqueIndex;not null"`
    Username     string         `gorm:"size:100;not null"`
    PasswordHash string         `gorm:"size:255"`
    GitHubID          *int64  `gorm:"column:github_id;index"`
    GitHubLogin       *string `gorm:"column:github_login;size:100"`
    AvatarURL         *string `gorm:"column:avatar_url;size:500"`
    GitHubAccessToken *string `gorm:"column:github_access_token;size:255"` // Токен доступа к GitHub API
    IsActive          bool    `gorm:"default:true"`
    CreatedAt    time.Time      `gorm:"autoCreateTime"`
    UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
    DeletedAt    gorm.DeletedAt `gorm:"index"` // Soft Delete
    Reviews      []CodeReview   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
```

### CodeReview Entity

```go
// internal/entity/review.go
type CodeReview struct {
    ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID         uuid.UUID       `gorm:"type:uuid;index;not null"`
    User           *User           `gorm:"foreignKey:UserID"`
    Title          string          `gorm:"size:255;not null"`
    Code           string          `gorm:"type:text;not null"`
    Language       string          `gorm:"size:50;not null"`
    Status         ReviewStatus    `gorm:"size:20;default:'pending'"`
    Result         *string         `gorm:"type:text"`
    CustomPrompt   *string         `gorm:"type:text"` // Пользовательский промпт
    CreatedAt      time.Time       `gorm:"autoCreateTime"`
    UpdatedAt      time.Time       `gorm:"autoUpdateTime"`
    CompletedAt    *time.Time
    DeletedAt      gorm.DeletedAt  `gorm:"index"`
    SecurityIssues []SecurityIssue `gorm:"foreignKey:ReviewID;constraint:OnDelete:CASCADE"`
}
```

### SecurityIssue Entity

```go
type SecurityIssue struct {
    ID          uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    ReviewID    uuid.UUID        `gorm:"type:uuid;index;not null"`
    Review      *CodeReview      `gorm:"foreignKey:ReviewID"`
    Severity    SecuritySeverity `gorm:"size:20;not null"`
    Title       string           `gorm:"size:255;not null"`
    Description string           `gorm:"type:text;not null"`
    LineStart   *int
    LineEnd     *int
    Suggestion  string           `gorm:"type:text"`
    CWE         *string          `gorm:"size:20"`
    CreatedAt   time.Time        `gorm:"autoCreateTime"`
    DeletedAt   gorm.DeletedAt   `gorm:"index"`
}
```

### GitHubInstallation Entity

```go
// internal/entity/github_installation.go
type GitHubInstallation struct {
    ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    InstallationID int64          `gorm:"index;unique"` // ID установки от GitHub
    AccountID      int64          `gorm:"index"`        // ID пользователя/организации GitHub
    AccountLogin   string         `gorm:"size:255"`
    AccountType    string         `gorm:"size:50"`      // "User" или "Organization"
    UserID         *uuid.UUID     `gorm:"type:uuid;index"` // Связь с локальным пользователем
    User           *User          `gorm:"foreignKey:UserID"`
    CreatedAt      time.Time      `gorm:"autoCreateTime"`
    UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
    DeletedAt      gorm.DeletedAt `gorm:"index"`
}
```

## Таблицы (SQL)

### github_installations

Таблица установок GitHub App.

```sql
CREATE TABLE github_installations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    installation_id BIGINT UNIQUE,
    account_id BIGINT,
    account_login VARCHAR(255),
    account_type VARCHAR(50),
    user_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_github_installations_installation_id ON github_installations(installation_id);
CREATE INDEX idx_github_installations_account_id ON github_installations(account_id);
```

### users

Таблица пользователей системы.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255),
    github_id BIGINT,
    github_login VARCHAR(100),
    avatar_url VARCHAR(500),
    github_access_token VARCHAR(255),
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

Миграции выполняются **автоматически** при запуске приложения через GORM `AutoMigrate()`:

```go
// internal/database/database.go
func (d *Database) AutoMigrate() error {
    return d.DB.AutoMigrate(
        &entity.User{},
        &entity.CodeReview{},
        &entity.SecurityIssue{},
    )
}
```

Это аналог `synchronize: true` в TypeORM — GORM автоматически создаёт таблицы и обновляет схему на основе Entity.

### Soft Delete

Все Entity поддерживают soft delete через `gorm.DeletedAt`:

```go
// Soft delete — запись не удаляется, а помечается deleted_at
repo.Delete(ctx, id)

// Hard delete — полное удаление
repo.HardDelete(ctx, id)

// GORM автоматически фильтрует удалённые записи
repo.FindAll(ctx)  // WHERE deleted_at IS NULL
```

### Подключение к БД

```go
// cmd/api/main.go
db, err := database.NewDatabase(cfg.Database.URL)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}

// Auto migrate
if err := db.AutoMigrate(); err != nil {
    log.Fatalf("Failed to migrate: %v", err)
}
```
