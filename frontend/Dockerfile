# Этап сборки (Build stage)
FROM node:18-alpine as build-stage

WORKDIR /app

# Копируем файлы package.json и устанавливаем зависимости
COPY package*.json ./
RUN npm install

# Копируем исходный код и собираем проект
COPY . .
RUN npm run build

# Этап запуска (Production stage)
FROM nginx:stable-alpine as production-stage

# Копируем собранные файлы из этапа build-stage в папку nginx
COPY --from=build-stage /app/dist /usr/share/nginx/html

# Копируем конфигурацию Nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Открываем порт 80 (Railway подставит $PORT, но nginx слушает 80 внутри контейнера по умолчанию, исправим в конфиге)
EXPOSE 80

# Запускаем Nginx (используем sed для подстановки PORT, если нужно, но по умолчанию Railway проксирует на порт контейнера)
# Обычно для Nginx на Railway удобно использовать порт из переменной окружения, но стандартный nginx слушает фиксированный порт.
# Railway автоматически маппит внешний порт на EXPOSEd порт.
CMD ["nginx", "-g", "daemon off;"]
