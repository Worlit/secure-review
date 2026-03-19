# Secure Review API

Backend сервис для анализа кода на безопасность и code review с использованием GitHub Copilot.

## Содержание

- [Архитектура](#архитектура)
- [Установка](#установка)
- [Тестирование](#тестирование)
- [Конфигурация](#конфигурация)
- [API Документация](#api-документация)
- [SOLID Принципы](#solid-принципы)
- [GORM ORM](#gorm-orm)

## Архитектура

Проект построен с использованием Clean Architecture и SOLID принципов.
Работа с БД реализована через GORM ORM с TypeORM-подобным API.

```
secure-review/
├── cmd/
│   └── api/
│       └── main.go              # Точка входа приложения
├── internal/
│   ├── config/
│   │   └── config.go            # Конфигурация приложения
│   ├── database/
│   │   └── database.go          # GORM подключение (аналог TypeORM DataSource)
│   ├── domain/
│   │   ├── errors.go            # Доменные ошибки
│   │   ├── repository.go        # Интерфейсы репозиториев
│   │   ├── review.go            # Domain модели для code review
│   │   ├── service.go           # Интерфейсы сервисов
│   │   └── user.go              # Domain модели пользователя
│   ├── entity/
│   │   ├── user.go              # GORM Entity User (аналог @Entity)
│   │   └── review.go            # GORM Entity CodeReview, SecurityIssue
│   ├── handler/
│   │   ├── auth_handler.go      # Обработчики авторизации
│   │   ├── github_handler.go    # Обработчики GitHub OAuth
│   │   ├── health_handler.go    # Обработчики health check
│   │   ├── review_handler.go    # Обработчики code review
│   │   └── user_handler.go      # Обработчики пользователя
│   ├── middleware/
│   │   ├── auth.go              # JWT аутентификация
│   │   └── logging.go           # Логирование
│   ├── repository/
│   │   ├── user_repository.go   # GORM репозиторий пользователей
│   │   ├── review_repository.go # GORM репозиторий code review
│   │   ├── user_adapter.go      # Адаптер для domain.UserRepository
│   │   └── review_adapter.go    # Адаптер для domain.ReviewRepository
│   ├── router/
│   │   └── router.go            # Настройка маршрутов
│   └── service/                 # Бизнес-логика
│       ├── auth/                # Аутентификация
│       │   ├── auth_service.go  # Сервис аутентификации
│       │   ├── jwt.go           # JWT токены
│       │   └── password.go      # Хэширование паролей
│       ├── user/                # Пользователи
│       │   └── user_service.go  # Сервис пользователей
│       ├── review/              # Code Review
│       │   └── review_service.go # Сервис code review
│       ├── github/              # GitHub интеграция
│       │   ├── github_auth_service.go  # GitHub OAuth сервис
│       │   └── github_app_service.go   # GitHub App сервис
│       ├── analyzer/            # AI анализ
│       │   └── copilot_analyzer.go # GitHub Copilot интеграция
│       └── pdf/                 # PDF отчёты
│           └── pdf_service.go   # Генерация PDF
├── tests/                       # Тесты
│   ├── app_test.go              # Интеграционные тесты
│   └── fakes/                   # Фейковые реализации
└── docs/                        # Документация
```

## Установка

### Требования

- Go 1.24+
- PostgreSQL 14+
- GitHub Copilot API Key
- GitHub OAuth App (optional)

### Шаги установки

1. Клонируйте репозиторий:

```bash
git clone https://github.com/yourusername/secure-review.git
cd secure-review
```

2. Скопируйте пример конфигурации:

```bash
cp .env.example .env
```

3. Отредактируйте `.env` файл с вашими настройками

4. Установите зависимости:

```bash
go mod download
```

5. Запустите приложение:

```bash
go run cmd/api/main.go
```

6. Запустите тесты:

```bash
go test -v ./...
```

## Тестирование

Проект использует unit и интеграционные тесты с фейковыми реализациями зависимостей.

### Запуск тестов

```bash
# Все тесты
go test -v ./...

# Unit тесты сервисов
go test -v ./internal/service/...

# Интеграционные тесты
go test -v ./tests/...

# С покрытием
go test -cover ./...
```

### Структура тестов

| Компонент | Файл тестов | Покрытие |
|-----------|-------------|----------|
| Auth Service | `internal/service/auth/auth_service_test.go` | Регистрация, логин, токены, смена пароля |
| JWT Token | `internal/service/auth/jwt_test.go` | Генерация, валидация, истекшие токены |
| Password Hasher | `internal/service/auth/password_test.go` | Хэширование, сравнение |
| User Service | `internal/service/user/user_service_test.go` | CRUD операции |
| Review Service | `internal/service/review/review_service_test.go` | Создание, анализ, GitHub интеграция |
| Health Handler | `internal/handler/health_handler_test.go` | Health, Ready endpoints |
| Auth Middleware | `internal/middleware/auth_test.go` | JWT валидация |
| Интеграционные | `tests/app_test.go` | End-to-end сценарии |

### Фейковые реализации

Фейки находятся в `tests/fakes/` и реализуют доменные интерфейсы:

- `FakeUserRepository` — in-memory хранилище пользователей
- `FakeReviewRepository` — in-memory хранилище reviews
- `FakeCodeAnalyzer` — mock AI анализатора
- `FakeGitHubService` — mock GitHub API

## Конфигурация

Все переменные окружения описаны в файле `.env.example`:

| Переменная             | Описание                     | Значение по умолчанию   |
| ---------------------- | ---------------------------- | ----------------------- |
| `SERVER_PORT`          | Порт сервера                 | `8080`                  |
| `SERVER_HOST`          | Хост сервера                 | `0.0.0.0`               |
| `GIN_MODE`             | Режим Gin (debug/release)    | `debug`                 |
| `DATABASE_URL`         | PostgreSQL connection string | -                       |
| `JWT_SECRET`           | Секретный ключ для JWT       | -                       |
| `JWT_EXPIRATION_HOURS` | Время жизни токена (часы)    | `24`                    |
| `COPILOT_API_KEY`      | API ключ GitHub Copilot     | -                       |
| `COPILOT_MODEL`        | Модель Copilot              | `gpt-4o`                |
| `GITHUB_CLIENT_ID`     | GitHub OAuth Client ID       | -                       |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth Client Secret   | -                       |
| `GITHUB_REDIRECT_URL`  | URL для callback             | -                       |
| `FRONTEND_URL`         | URL фронтенда                | `http://localhost:3000` |

## API Документация

Подробная API документация находится в [docs/API.md](./API.md).

### Основные эндпоинты

#### Аутентификация

- `POST /api/v1/auth/register` - Регистрация
- `POST /api/v1/auth/login` - Вход
- `POST /api/v1/auth/refresh` - Обновление токена
- `POST /api/v1/auth/change-password` - Смена пароля

#### GitHub OAuth

- `GET /api/v1/auth/github` - Получить URL для OAuth
- `GET /api/v1/auth/github/callback` - Callback (редирект)
- `POST /api/v1/auth/github/callback/json` - Callback (JSON)
- `POST /api/v1/auth/github/link` - Привязать GitHub аккаунт
- `POST /api/v1/auth/github/unlink` - Отвязать GitHub аккаунт

#### Пользователи

- `GET /api/v1/users/me` - Профиль пользователя
- `PUT /api/v1/users/me` - Обновить профиль
- `DELETE /api/v1/users/me` - Удалить аккаунт

#### Code Review

- `POST /api/v1/reviews` - Создать review
- `GET /api/v1/reviews` - Список reviews
- `GET /api/v1/reviews/:id` - Получить review
- `DELETE /api/v1/reviews/:id` - Удалить review
- `POST /api/v1/reviews/:id/reanalyze` - Повторный анализ

#### Health Check

- `GET /health` - Проверка здоровья
- `GET /ready` - Готовность

## SOLID Принципы

### Single Responsibility Principle (SRP)

Каждый компонент имеет одну ответственность:

- `AuthService` - только аутентификация
- `UserService` - только операции с пользователями
- `ReviewService` - только операции с reviews
- `CopilotAnalyzer` - только анализ кода

### Open/Closed Principle (OCP)

Система открыта для расширения через интерфейсы:

- `CodeAnalyzer` - можно добавить другие анализаторы (не только Copilot)
- `TokenGenerator` - можно заменить JWT на другую систему
- `PasswordHasher` - можно использовать другой алгоритм хэширования

### Liskov Substitution Principle (LSP)

Все реализации полностью соответствуют интерфейсам:

- `UserRepositoryAdapter` implements `domain.UserRepository`
- `ReviewRepositoryAdapter` implements `domain.ReviewRepository`
- `CopilotAnalyzer` implements `CodeAnalyzer`

### Interface Segregation Principle (ISP)

Интерфейсы разделены по назначению:

- `UserRepository` - операции с пользователями в БД
- `ReviewRepository` - операции с reviews в БД
- `AuthService` - аутентификация
- `GitHubAuthService` - GitHub OAuth

### Dependency Inversion Principle (DIP)

Зависимости инжектируются через конструкторы:

- Сервисы зависят от интерфейсов репозиториев
- Handlers зависят от интерфейсов сервисов
- Легкое тестирование через моки

## GORM ORM

Работа с базой данных реализована через GORM — Go ORM с TypeORM-подобным API.

### Entity Layer

Entity определяются с GORM тегами (аналог декораторов TypeORM):

```go
// internal/entity/user.go
type User struct {
    // @PrimaryGeneratedColumn("uuid")
    ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    
    // @Column({ unique: true })
    Email string `gorm:"size:255;uniqueIndex;not null"`
    
    // @Column()
    Username string `gorm:"size:100;not null"`
    
    // @DeleteDateColumn() — Soft Delete
    DeletedAt gorm.DeletedAt `gorm:"index"`
    
    // @OneToMany(() => CodeReview, review => review.user)
    Reviews []CodeReview `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
```

### Database Connection

```go
// internal/database/database.go
// Аналог new DataSource({...}).initialize()
db, err := database.NewDatabase(databaseURL)

// Аналог synchronize: true
db.AutoMigrate()

// Аналог manager.transaction()
db.Transaction(func(tx *gorm.DB) error {
    // ...
})
```

### Repository Pattern

```go
// GORM методы (аналоги TypeORM)
repo.FindByID(ctx, id)           // findOne({ where: { id } })
repo.FindByIDWithReviews(ctx, id) // { relations: ['reviews'] }
repo.Create(ctx, &user)          // save(user)
repo.Delete(ctx, id)             // softDelete(id)
repo.UpdateFields(ctx, id, map)  // update(id, { ...fields })
```

### Преимущества GORM

| TypeORM | GORM |
|---------|------|
| `@Entity()` | `gorm:"..."` теги |
| `@Column()` | `gorm:"size:255;not null"` |
| `@PrimaryGeneratedColumn("uuid")` | `gorm:"type:uuid;primaryKey"` |
| `@CreateDateColumn()` | `gorm:"autoCreateTime"` |
| `@DeleteDateColumn()` | `gorm.DeletedAt` |
| `@ManyToOne()` | `gorm:"foreignKey:..."` |
| `{ relations: [...] }` | `.Preload("...")` |
| `synchronize: true` | `AutoMigrate()` |
