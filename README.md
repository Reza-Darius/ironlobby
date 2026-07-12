# Iron Lobby

A small application for Heart of Iron 4 to help people organize multiplayer lobbies!

Run with `docker compose up`

`.env` format for configuration:

```
# addr the app listens on inside the container
PORT=3000

MIGRATION_PATH=./migrations

POSTGRES_USER=schnib-user
POSTGRES_PASSWORD=mypassword
POSTGRES_DB=schnib-db
```
