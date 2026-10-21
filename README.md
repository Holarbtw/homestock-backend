# Homestock Backend

Бэкенд мобильного приложения для учёта домашних запасов.  
Построен на микросервисной архитектуре (Go, PostgreSQL, Nginx Gateway, Docker Compose).

---

## Быстрый старт

1. Скопируйте файл с переменными окружения:
   ```bash
   cp .env.example .env
   ```

2. Запустите проект через Docker Compose:
   ```bash
   docker compose up --build -d
   ```

`AUTH_SERVICE_PORT` задаёт порт сервиса внутри Docker и адрес, к которому обращается Nginx.
Имена баз данных из `.env` используются при первой инициализации тома PostgreSQL.
Если том уже существует, изменение имени базы в `.env` не создаст новую базу автоматически:
её нужно создать отдельно и применить миграции.

---

## Swagger UI

После запуска контейнеров документация и тестирование API доступны по адресу:

**[http://localhost/api/v1/auth/swagger/index.html](http://localhost/api/v1/auth/swagger/index.html)** *(через Nginx Gateway)*  

Для вызова защищенных эндпоинтов используйте кнопку **Authorize** 🔓 в Swagger (передавать как `Bearer <access_token>`).

---

## Полезные команды

- **Логи сервиса**: `docker compose logs -f auth-service`
- **Остановить контейнеры**: `docker compose down`
- **Остановить с очисткой базы**: `docker compose down -v`
