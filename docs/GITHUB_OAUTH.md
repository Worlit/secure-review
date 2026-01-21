# Настройка GitHub OAuth

Инструкция по настройке GitHub OAuth для приложения Secure Review.

## Шаг 1: Создание OAuth приложения в GitHub

1. Перейдите в [GitHub Developer Settings](https://github.com/settings/developers)

2. Нажмите **"New OAuth App"**

3. Заполните форму:

   | Поле                           | Значение                                            |
   | ------------------------------ | --------------------------------------------------- |
   | **Application name**           | Secure Review (или любое название)                  |
   | **Homepage URL**               | `http://localhost:3000` (или URL вашего фронтенда)  |
   | **Application description**    | Code security analysis and review tool              |
   | **Authorization callback URL** | `http://localhost:8080/api/v1/auth/github/callback` |

4. Нажмите **"Register application"**

5. Скопируйте **Client ID**

6. Нажмите **"Generate a new client secret"** и скопируйте **Client Secret**

## Шаг 2: Настройка переменных окружения

Добавьте полученные значения в файл `.env`:

```env
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_client_secret_here
GITHUB_REDIRECT_URL=http://localhost:8080/api/v1/auth/github/callback
```

## Шаг 3: Настройка для продакшена

Для продакшена измените URL на реальные:

```env
GITHUB_REDIRECT_URL=https://api.yourdomain.com/api/v1/auth/github/callback
FRONTEND_URL=https://yourdomain.com
```

И обновите настройки OAuth приложения в GitHub:

- **Homepage URL**: `https://yourdomain.com`
- **Authorization callback URL**: `https://api.yourdomain.com/api/v1/auth/github/callback`

## Использование OAuth

### Процесс авторизации

1. Фронтенд запрашивает URL для OAuth:

   ```
   GET /api/v1/auth/github
   ```

   _Запрашиваемые права (scopes):_ `user:email`, `read:user`, `repo` (для чтения списка репозиториев)

2. Пользователь перенаправляется на GitHub для авторизации

3. После авторизации GitHub перенаправляет на callback URL

4. Backend обменивает код на токен и создаёт/аутентифицирует пользователя

5. Пользователь перенаправляется на фронтенд с JWT токеном

### Пример фронтенд интеграции

```javascript
// Получить URL для OAuth
const response = await fetch('/api/v1/auth/github');
const { url } = await response.json();

// Перенаправить пользователя
window.location.href = url;
```

### Обработка callback на фронтенде

```javascript
// На странице /auth/callback
const params = new URLSearchParams(window.location.search);
const token = params.get('token');

if (token) {
  localStorage.setItem('token', token);
  // Перенаправить в приложение
  window.location.href = '/dashboard';
}
```

## Привязка GitHub к существующему аккаунту

Если пользователь зарегистрировался через email/password и хочет привязать GitHub:

```javascript
// Получить URL для OAuth (аналогично)
const response = await fetch('/api/v1/auth/github');
const { url, state } = await response.json();

// Сохранить state для проверки
sessionStorage.setItem('oauth_state', state);

// Перенаправить
window.location.href = url;
```

После возврата, отправить POST запрос:

```javascript
const response = await fetch('/api/v1/auth/github/link', {
  method: 'POST',
  headers: {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ code: params.get('code') }),
});
```

## Безопасность

- Всегда проверяйте `state` параметр для защиты от CSRF
- Храните `client_secret` только на сервере
- Используйте HTTPS в продакшене
- Ограничьте OAuth scopes только необходимыми (`user:email`, `read:user`)
