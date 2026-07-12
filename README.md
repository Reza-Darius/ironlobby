# Schnibbelbox

A small little project to play around with docker and postgres

Run with `docker compose up`

`.env` format for configuration:

```
# addr the app listens on inside the container
PORT=3000

MIGRATION_PATH=./migrations

# for SQLite
DB_PATH=./database.db

POSTGRES_USER=schnib-user
POSTGRES_PASSWORD=mypassword
POSTGRES_DB=schnib-db
```
