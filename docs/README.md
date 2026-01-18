# Secure Review API

Backend сервис для анализа кода на безопасность и code review с использованием OpenAI.

## Содержание

- [Архитектура](#архитектура)
- [Установка](#установка)
- [Конфигурация](#конфигурация)
- [API Документация](#api-документация)
- [SOLID Принципы](#solid-принципы)

## Архитектура

Проект построен с использованием Clean Architecture и SOLID принципов:

```
secure-review/
├── cmd/
│   └── api/
│       └── main.go              # Точка входа приложения
├── internal/
│   ├── config/
│   │   └── config.go            # Конфигурация приложения
│   ├── domain/
│   │   ├── errors.go            # Доменные ошибки
│   │   ├── repository.go        # Интерфейсы репозиториев
│   │   ├── review.go            # Модели для code review
│   │   ├── service.go           # Интерфейсы сервисов
│   │   └── user.go              # Модели пользователя
│   ├── handler/
│   │   ├── auth_handler.go      # Обработчики авторизации
│   │   ├── github_handler.go    # Обработчики GitHub OAuth
│   │   ├── health_handler.go    # Обработчики health check
│   │   ├── review_handler.go    # Обработчики code review
│   │   ├── types.go             # DTO типы
│   │   └── user_handler.go      # Обработчики пользователя
│   ├── middleware/
│   │   ├── auth.go              # JWT аутентификация
│   │   ├── cors.go              # CORS middleware
│   │   └── logging.go           # Логирование
│   ├── repository/
│   │   ├── postgres.go          # Подключение к PostgreSQL
│   │   ├── review_repository.go # Репозиторий code review
│   │   └── user_repository.go   # Репозиторий пользователей
│   ├── router/
│   │   └── router.go            # Настройка маршрутов
│   └── service/
│       ├── auth_service.go      # Сервис аутентификации
│       ├── github_auth_service.go # GitHub OAuth сервис
│       ├── jwt.go               # JWT токены
│       ├── openai_analyzer.go   # OpenAI интеграция
│       ├── password.go          # Хэширование паролей
│       ├── review_service.go    # Сервис code review
│       └── user_service.go      # Сервис пользователей
└── docs/                        # Документация
```

## Установка

### Требования

- Go 1.21+
- PostgreSQL 14+
- OpenAI API Key
- GitHub OAuth App

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
| `OPENAI_API_KEY`       | API ключ OpenAI              | -                       |
| `OPENAI_MODEL`         | Модель OpenAI                | `gpt-4`                 |
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
- `OpenAICodeAnalyzer` - только анализ кода

### Open/Closed Principle (OCP)

Система открыта для расширения через интерфейсы:

- `CodeAnalyzer` - можно добавить другие анализаторы (не только OpenAI)
- `TokenGenerator` - можно заменить JWT на другую систему
- `PasswordHasher` - можно использовать другой алгоритм хэширования

### Liskov Substitution Principle (LSP)

Все реализации полностью соответствуют интерфейсам:

- `PostgresUserRepository` implements `UserRepository`
- `PostgresReviewRepository` implements `ReviewRepository`
- `OpenAICodeAnalyzer` implements `CodeAnalyzer`

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
