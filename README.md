# Organization Structure API

Go backend для управления древовидной структурой подразделений компании и сотрудниками. Проект показывает работу с REST API, PostgreSQL, транзакциями, recursive CTE, Docker и разделением бизнес-логики по слоям.

> Учебный backend-проект с акцентом на чистую архитектуру, корректную работу с иерархическими данными и предсказуемое поведение API.

## Tech stack

- **Go 1.24**, стандартная библиотека `net/http`
- **PostgreSQL 16**
- **GORM**
- **Goose** migrations
- **Docker / Docker Compose**
- unit tests с mock-зависимостями

## Что реализовано

### Departments

- `POST /departments/` — создание подразделения;
- `GET /departments/{id}` — получение дерева с ограничиваемой глубиной;
- `PATCH /departments/{id}` — переименование и перемещение подразделения;
- `DELETE /departments/{id}` — удаление в режимах `cascade` и `reassign`.

Для перемещения подразделений используется рекурсивная проверка дерева через **CTE**, которая запрещает циклы: департамент нельзя переместить внутрь самого себя или собственного поддерева.

Операции, затрагивающие несколько сущностей, выполняются в **PostgreSQL-транзакциях**.

### Employees

- `POST /departments/{id}/employees/` — создание сотрудника в подразделении;
- проверка существования подразделения;
- валидация входных данных;
- корректное отображение доменных ошибок в HTTP status codes.

## Architecture

```text
cmd/api/              # entrypoint, dependency wiring, graceful shutdown
config/               # environment configuration, fail-fast checks
internal/domain/       # entities and interfaces
internal/usecase/      # business rules
internal/repository/   # PostgreSQL/GORM, transactions, recursive CTE
internal/handler/      # HTTP transport, JSON parsing, error mapping
migrations/            # SQL migrations
```

Бизнес-логика отделена от HTTP и persistence-слоя. Зависимости собираются в entrypoint, а доменный слой не зависит от ORM.

## Быстрый запуск

```bash
git clone https://github.com/BArL43/org-structure-api.git
cd org-structure-api
docker compose up --build
```

При старте:

1. поднимается PostgreSQL;
2. healthcheck ожидает готовность БД;
3. Goose применяет миграции;
4. Go-приложение собирается multi-stage Docker build и запускается на `:8080`.

Проверка:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{"status":"OK","message":"Server is running"}
```

## Tests

```bash
go test ./...
```

Тесты проверяют HTTP/handler-слой и его взаимодействие с бизнес-логикой через изолированные mock-зависимости.

## Почему этот проект полезен как backend-кейс

В нём есть не только CRUD: проект требует корректной работы с иерархическими данными, предотвращения циклов, транзакционного изменения структуры и явного разделения transport / business / persistence слоёв.
