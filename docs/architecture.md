# Архитектура проекта

Проект построен как монорепозиторий, содержащий backend (Python/FastAPI) и frontend (Vue/TypeScript).

## Высокоуровневая схема

```mermaid
graph TD
    User[Пользователь] -->|Browser| UI[Frontend (Vue 3)]
    UI -->|REST API| API[Backend (FastAPI)]

    subgraph Backend
        API --> Auth[Auth Service]
        API --> Analyzer[Analysis Service]
        Analyzer -->|Validation| Models[Pydantic Models]
        Auth -->|DB Access| DB[PostgreSQL]
        Analyzer -->|Context| DB
    end

    Analyzer -->|Inference| OpenAI[OpenAI API]
```

## Backend (FastAPI)

Бэкенд организован по принципам Clean Architecture / Layered Architecture для обеспечения модульности и тестируемости.

### Структура директорий

- `api/`: Роутеры и контроллеры (endpoints).
  - `auth.py`: Регистрация и аутентификация.
  - `router.py`: Основной маршрутизатор.
  - `deps.py`: Внедрение зависимостей (Dependency Injection).
- `core/`: Глобальные настройки и утилиты.
  - `config.py`: Управление конфигурацией (env vars).
  - `database.py`: Подключение к БД.
  - `security.py`: Хеширование паролей.
- `models/`: Модели данных.
  - `schemas.py`: Pydantic схемы (DTO) для API.
  - `sql_models.py`: SQLAlchemy модели (ORM) для БД.
- `services/`: Бизнес-логика.
  - `interfaces.py`: Абстракции (ICodeAnalyzer).
  - `openai_service.py`: Реализация анализа через OpenAI.

### Принципы

- **SOLID**: Использование интерфейсов (`ICodeAnalyzer`) позволяет легко заменять реализации анализаторов (например, добавить локальный SAST вместо OpenAI).
- **Dependency Injection**: Зависимости (БД, сервисы) внедряются в handler-функции через `Depends`.
- **Async**: Полностью асинхронный I/O (DB, HTTP calls).

## Frontend (Vue 3)

- **Framework**: Vue 3 (Composition API, `<script setup>`).
- **Language**: TypeScript.
- **Build Tool**: Vite.
- **State Management**: (Планируется Pinia).
- **Routing**: Vue Router.

## Интеграции

- **OpenAI API**: Используется модель GPT-4o/GPT-3.5-turbo для семантического анализа кода, объяснения уязвимостей и генерации фиксов. Ответы парсятся как JSON для структурированного отображения.
- **GitHub**: (В планах) OAuth авторизация и доступ к репозиториям.
