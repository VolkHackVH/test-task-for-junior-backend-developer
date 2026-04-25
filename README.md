# Task Service

HTTP API для управления задачами и настройками их повторяемости.

## Что реализовано

- Базовый CRUD задач (`POST/GET/PUT/DELETE /api/v1/tasks`).
- Настройки повторяемости как часть payload задачи (`recurrence` в `POST/PUT /tasks`).
- Поддерживаемые типы повторяемости:
  - `daily` (каждый `n`-й день),
  - `monthly` (день месяца `1..30`),
  - `specific_dates` (список конкретных дат),
  - `odd_days` / `even_days`.
- Отдельные таблицы для правил повторяемости:
  - `task_recurrences`,
  - `specific_dates`.
- Материализация задач по правилу повторяемости:
  - при создании/обновлении recurrence сервис генерирует экземпляры задач на 30 дней вперед,
  - дубли предотвращаются уникальным индексом `(source_task_id, scheduled_for)`.

## Принятые допущения

- Повторяемость хранится отдельно от основной задачи, но задается в одном запросе `POST/PUT /tasks`.
- `recurrence` опционален: задача может быть без повторяемости.
- Для `daily`, `monthly`, `odd_days`, `even_days`, `specific_dates` поля взаимоисключающие и валидируются в usecase.
- При обновлении recurrence будущие сгенерированные экземпляры пересоздаются (начиная с текущей даты).
- Предел генерации экземпляров фиксированный: `30` дней.
- Если recurrence неактивен (`is_active=false`), будущие экземпляры удаляются и новые не генерируются.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск

```bash
docker compose down -v
docker compose up --build
```

Почему нужен `down -v`: SQL из `migrations` применяется только при инициализации пустого volume.

## Миграции

В `docker-compose.yml` смонтирована вся папка `migrations` в `docker-entrypoint-initdb.d`, поэтому на чистом volume применяются все файлы:

- `0001_create_tasks.up.sql`
- `0002_create_task_recurrences.up.sql`
- `0003_create_specific_dates.up.sql`
- `0004_add_task_occurrences.up.sql`

## Swagger

- UI: `http://localhost:8080/swagger/`
- OpenAPI JSON: `http://localhost:8080/swagger/openapi.json`

## API

Базовый префикс: `/api/v1`

- `POST /tasks`
- `GET /tasks`
- `GET /tasks/{id}`
- `PUT /tasks/{id}`
- `DELETE /tasks/{id}`

### Пример валидного `daily` payload

```json
{
  "title": "Prepare release",
  "description": "Collect release notes and check migrations",
  "status": "new",
  "recurrence": {
    "recurrence_type": "daily",
    "interval_days": 1,
    "start_date": "2026-04-23",
    "is_active": true
  }
}
```

## Отдельная БД для интеграционных тестов

Для тестов используется отдельный compose-файл: `docker-compose.test.yml`.

Поднять test БД:

```bash
make test-db-up
```

Сбросить test БД (с удалением volume и повторным применением миграций):

```bash
make test-db-reset
```

Остановить test БД:

```bash
make test-db-down
```

## Интеграционные тесты repository

Запуск всех интеграционных тестов repository:

```bash
make test-repo
```

Запуск одного теста-примера:

```bash
make test-repo-one
```

По умолчанию используется DSN:

`postgres://postgres:postgres@localhost:5433/taskservice_test?sslmode=disable`

Его можно переопределить:

```bash
make test-repo TEST_DSN="postgres://postgres:postgres@localhost:5433/taskservice_test?sslmode=disable"
```
