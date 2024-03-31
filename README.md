# echosight

🚧 UNDER CONSTRUCTION 🚧

A simple monitoring system


## Development

**Dev Dependencies**

- [`go-task`](https://github.com/go-task/task)
- [`go-migrate`](https://github.com/golang-migrate/migrate)

**Dependencies** (see `docker-compose.yml`)

- Postgres
- InfluxDB
- Redis (optional)

Export `POSTGRESQL_URL` in your shell environment:

```bash
export POSTGRESQL_URL='postgres://root:root@localhost:5432/echosight_dev?sslmode=disable'
```

Start Postgres create database `echosight_dev` and run `task migrate:up`.

Create `config/debug-config.json` from `config/template-config.json` and customize `config/debug-config.json`.

Run `go run cmd/server/main.go --config config/debug-config.json` or use `.vscode/launch.json`

