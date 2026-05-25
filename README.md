### Тестовое задание Effective Mobile

###### Для запуска приложения:

- запустить базу данных с помощью docker-compose:
```bash
docker compose up db -d
```
- запустить миграции с помощью [goose](https://github.com/pressly/goose):
```bash
GOOSE_DRIVER=postgres GOOSE_MIGRATION_DIR=migrations GOOSE_DBSTRING=postgres://postgres:postgres@localhost:5527/postgres goose up
```
- запустить приложение в docker-compose:
```bash
docker-compose up app -d
```
---
Документация API доступна по адресу: http://localhost:8080/api/docs

P.S. структура директорий была упрощена намеренно.
