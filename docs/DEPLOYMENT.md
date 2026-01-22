# Деплой приложения

## Docker

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /secure-review ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /secure-review .

EXPOSE 8080

CMD ["./secure-review"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - '8080:8080'
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - JWT_SECRET=${JWT_SECRET}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - GITHUB_CLIENT_ID=${GITHUB_CLIENT_ID}
      - GITHUB_CLIENT_SECRET=${GITHUB_CLIENT_SECRET}
      - GITHUB_REDIRECT_URL=${GITHUB_REDIRECT_URL}
      - FRONTEND_URL=${FRONTEND_URL}
      - LOG_LEVEL=info
      - LOG_FORMAT=json
      - GIN_MODE=release
    depends_on:
      - postgres
    restart: unless-stopped

  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: secure_review
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - '5432:5432'

volumes:
  postgres_data:
```

## Запуск

```bash
# Локально
go run cmd/api/main.go

# С Docker
docker-compose up -d

# Билд
go build -o secure-review ./cmd/api
./secure-review
```

## Переменные окружения для продакшена

```env
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
GIN_MODE=release

# Database (используйте managed PostgreSQL)
DATABASE_URL=postgresql://user:pass@host:5432/dbname?sslmode=require

# JWT (сгенерируйте надёжный секрет)
JWT_SECRET=very-long-random-string-at-least-32-characters
JWT_EXPIRATION_HOURS=24

# OpenAI
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4

# GitHub OAuth (обновите URL на продакшен)
GITHUB_CLIENT_ID=...
GITHUB_CLIENT_SECRET=...
GITHUB_REDIRECT_URL=https://api.yourdomain.com/api/v1/auth/github/callback

# Frontend
FRONTEND_URL=https://yourdomain.com

# Logging
LOG_LEVEL=info # debug, info, warn, error
LOG_FORMAT=json # json, text
```

## Health Checks

Для мониторинга используйте endpoints:

- `GET /health` - Проверка здоровья
- `GET /ready` - Готовность к приёму трафика

## Рекомендации

1. **База данных**: Используйте managed PostgreSQL (Railway, Supabase, AWS RDS)
2. **Секреты**: Храните в переменных окружения или secret manager
3. **HTTPS**: Используйте reverse proxy (nginx, Caddy) с SSL
4. **Мониторинг**: Подключите логирование и метрики
5. **Бэкапы**: Настройте автоматические бэкапы БД
