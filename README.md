# Organization Structure API

![CI](https://github.com/BArL43/org-structure-api/actions/workflows/ci.yml/badge.svg)

Go backend для управления древовидной структурой подразделений компании и сотрудниками. Учебный проект с акцентом на REST API, PostgreSQL, транзакции, recursive CTE, `context.Context`, тестируемую бизнес-логику и воспроизводимый Docker-запуск.

Это основной Go-проект в моём публичном портфолио: здесь лучше всего видны работа со слоями backend-приложения, транзакциями, SQL-ограничениями, HTTP-контрактом и тестами.

## Tech stack

- **Go 1.25**, стандартная библиотека `net/http`
- **PostgreSQL 16**, GORM, recursive CTE
- **Goose** migrations
- **Docker / Docker Compose**
- **GitHub Actions**: tests, `go vet`, Compose validation и smoke test

## Что реализовано

### Departments

- `POST /departments/` — создание подразделения;
- `GET /departments/{id}?depth=1..5&include_employees=true|false` — получение дерева;
- `PATCH /departments/{id}` — частичное обновление имени и/или родителя;
- `parent_id: null` переносит подразделение в корень;
- `DELETE /departments/{id}?mode=cascade` — удаление ветки;
- `DELETE /departments/{id}?mode=reassign&reassign_to_department_id=...` — атомарный перенос сотрудников, подъём дочерних подразделений и удаление узла.

При перемещении выполняется recursive CTE-проверка, запрещающая циклы. Уникальность имени контролируется в рамках одного parent, включая root-уровень.

### Employees

- `POST /departments/{id}/employees/` — создание сотрудника;
- проверка существования подразделения;
- валидация имени и должности;
- сортировка сотрудников по имени при чтении дерева.

## Архитектура

```text
cmd/api/              # entrypoint, dependency wiring, healthcheck, graceful shutdown
config/               # environment configuration and validation
internal/domain/       # entities, errors and interfaces
internal/usecase/      # business rules
internal/repository/   # PostgreSQL/GORM, CTE and transactions
internal/handler/      # HTTP transport, strict JSON decoding, error mapping
migrations/            # PostgreSQL schema
```

`context.Context` передаётся от HTTP request через use case до GORM/SQL-запросов. PATCH хранит отдельно значение поля и факт его присутствия, поэтому обновление одного поля не затирает другое.

## Быстрый запуск

```bash
git clone https://github.com/BArL43/org-structure-api.git
cd org-structure-api
docker compose up --build
```

Compose поднимает PostgreSQL, ждёт healthcheck, применяет Goose-миграции и запускает API на `:8080`.

```bash
curl http://localhost:8080/health
```

```json
{"status":"ok"}
```

### Пример

```bash
curl -X POST http://localhost:8080/departments/ \
  -H 'Content-Type: application/json' \
  -d '{"name":"Backend"}'

curl 'http://localhost:8080/departments/1?depth=2&include_employees=true'
```

## Tests

```bash
go test -race ./...
go vet ./...
```

Unit-тесты покрывают handler/usecase сценарии: строгий JSON, query/path parsing, частичный PATCH, `parent_id: null`, duplicate detection, cycle detection и delete validation. CI дополнительно поднимает весь Docker Compose stack и выполняет smoke test API.
