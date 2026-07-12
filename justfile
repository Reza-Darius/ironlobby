set dotenv-load := true

run:
    go run ./cmd/*

migrate:
    goose up

dblogin:
  # we could alternatively expose a port on the pg container to connect to
  docker compose exec db psql -U {{env('POSTGRES_USER')}} -d {{env('POSTGRES_DB')}}
