# Cистема пропусков

Учебный проекn - веб‑приложение для управления пропусками на авто. Для пользователей в проекте существуют три роли: `admin`, `guard`, `resident`.

## Реализованные требования
- Аутентификация (JWT) и контроль доступа.
- Полноценный REST API с CRUD для пользователей, пропусков и гостевых заявок.
- PostgreSQL + миграции + типобезопасные запросы (sqlc).
- Логирование и метрики реализованы с помощью Prometheus.
- Swagger UI (`/docs`).

## Быстрый старт (Docker Compose)
```bash
cd deploy/compose

docker compose up --build
```

Сервисы:
- Backend: http://localhost:8080
- Swagger UI: http://localhost:8080/docs
- Frontend: http://localhost:5173
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

Grafana автоматически подхватывает дашборд `Project Overview` при старте:
- Dashboard: `Project Overview`

При старте создаётся администратор:
- email: `admin@example.com`
- password: `admin123`

Эти значения задаются через переменные окружения:
- `BOOTSTRAP_ADMIN_EMAIL`
- `BOOTSTRAP_ADMIN_PASSWORD`
- `BOOTSTRAP_ADMIN_NAME`

## Локальный запуск (без Docker)
1) Поднять Postgres и настроить `DB_DSN`.
2) Запустить сервер:
```bash
go run ./cmd/api
```
3) Запустить фронтенд:
```bash
cd frontend
npm ci
npm run dev
```

## Переменные окружения backend
- `APP_ENV` (`dev`/`prod`)
- `HTTP_ADDR` (default `:8080`)
- `DB_DSN` (например `postgres://postgres:postgres@localhost:5432/pipo?sslmode=disable`)
- `JWT_SECRET`
- `JWT_REFRESH_SECRET`
- `ACCESS_TTL` (например `15m`)
- `REFRESH_TTL` (например `168h`)
- `MIGRATE_ON_START` (`true/false`)
- `CORS_ORIGINS` (CSV)
- `BOOTSTRAP_ADMIN_EMAIL`
- `BOOTSTRAP_ADMIN_PASSWORD`
- `BOOTSTRAP_ADMIN_NAME`

## SQLC и миграции
- Миграции: `db/migrations/`
- Запросы: `db/queries/`
- Конфиг sqlc: `sqlc.yaml`

Генерация sqlc:
```bash
./scripts/sqlc_generate.sh
```

## Тесты
Unit:
```bash
go test ./internal/service -run 'TestServiceUnit|TestValidate'
go test ./internal/http
```

Functional / integration (требуется Docker):
```bash
go test -tags=integration ./internal/service ./internal/http
```

Полная команда:
```bash
go test ./...
go test -tags=integration ./internal/service ./internal/http
```

Пороговые значения покрытия в CI:
- `internal/service` unit: не ниже 85%
- `internal/http` combined (unit + integration): не ниже 80%

## CI/CD артефакт сборки
GitHub Actions формирует runtime-артефакт backend для `push` в `main` и `tags`:
- `pipo-backend-image-<sha>.tar.gz`
- `pipo-backend-image-<sha>.sha256`

Поведение workflow:
- `pull_request`: проверки Go/frontend + проверка сборки backend docker image.
- `push main`: проверки + smoke + публикация runtime-артефакта в Actions artifacts.
- `push tag`: проверки + smoke + публикация runtime-артефакта в Actions artifacts.

## Kubernetes
Манифесты находятся в `deploy/k8s/`. Включают:
- Backend + Frontend deployments
- PostgreSQL StatefulSet
- Ingress
- Prometheus + Grafana
- Loki + Promtail

## Роли и доступ
- `admin`: управление пользователями, пропусками, гостевыми заявками.
- `guard`: поиск пропусков, отметка въезда/выезда.
- `resident`: управление собственными пропусками и гостевыми заявками (обязателен `plot_number`/«участок»).
