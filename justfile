# Iron Lobby dev tasks. Run `just` to see them all.
#
#   just up       start everything (api + db in docker, frontend in background)
#   just down     stop everything
#   just dev      like `up`, but the frontend runs in this terminal (Ctrl+C stops it)

set dotenv-load := true
set shell := ["bash", "-euo", "pipefail", "-c"]

run_dir := ".run"
fe_pid := run_dir / "frontend.pid"
fe_log := run_dir / "frontend.log"
fe_port := "3001"
api_port := env("PORT", "3000")

# list recipes
default:
    @just --list --unsorted

# start api, db and frontend (frontend in the background)
up: backend frontend
    @echo ""
    @echo "  Iron Lobby is up → http://localhost:{{fe_port}}"
    @echo "  api: http://localhost:{{api_port}}/api   frontend log: {{fe_log}}"
    @echo "  stop with: just down"

# stop everything (data in the db volume is kept)
down: frontend-stop
    docker compose down

# restart everything, rebuilding the api image
restart: down up

# start api + db, frontend in the foreground (Ctrl+C stops the frontend only)
dev: backend _deps
    cd frontend && ./node_modules/.bin/vite --port {{fe_port}} --strictPort

# what's running
status:
    @docker compose ps
    @if [ -f {{fe_pid}} ] && kill -0 "$(cat {{fe_pid}})" 2>/dev/null; then \
        echo "frontend: running (pid $(cat {{fe_pid}})) → http://localhost:{{fe_port}}"; \
     else echo "frontend: stopped"; fi

# follow logs: `just logs` (api), `just logs db`, `just logs frontend`
logs service="api":
    @if [ "{{service}}" = "frontend" ]; then tail -f {{fe_log}}; else docker compose logs -f {{service}}; fi

# ---- pieces ------------------------------------------------------------------

# build the api image and start api + db, then wait until the api answers
backend:
    docker compose up -d --build
    @echo -n "waiting for api"
    @for i in $(seq 1 30); do \
        curl -sf -o /dev/null localhost:{{api_port}}/api/health && { echo " ok"; exit 0; }; \
        echo -n "."; sleep 1; \
     done; echo " timed out (see: just logs)"; exit 1

# start the frontend dev server in the background
frontend: _deps
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p {{run_dir}}
    if [ -f {{fe_pid}} ] && kill -0 "$(cat {{fe_pid}})" 2>/dev/null; then
        echo "frontend already running (pid $(cat {{fe_pid}}))"; exit 0
    fi
    cd frontend
    nohup ./node_modules/.bin/vite --port {{fe_port}} --strictPort > ../{{fe_log}} 2>&1 &
    echo $! > ../{{fe_pid}}
    # give vite a moment; fail loudly if it died (e.g. port already taken)
    for i in $(seq 1 20); do
        curl -sf -o /dev/null localhost:{{fe_port}} && { echo "frontend started (pid $(cat ../{{fe_pid}}))"; exit 0; }
        kill -0 "$(cat ../{{fe_pid}})" 2>/dev/null || break
        sleep 0.5
    done
    echo "frontend failed to start:"; tail -n 20 ../{{fe_log}}; rm -f ../{{fe_pid}}; exit 1

# stop the background frontend dev server
frontend-stop:
    @if [ -f {{fe_pid}} ] && kill "$(cat {{fe_pid}})" 2>/dev/null; then \
        echo "frontend stopped"; else echo "frontend not running"; fi
    @rm -f {{fe_pid}}

# production build of the frontend into frontend/dist
build:
    cd frontend && npm run build

# type-check the frontend
check:
    cd frontend && npm run check

# DESTRUCTIVE: stop everything and delete the database volume
reset-db:
    #!/usr/bin/env bash
    read -rp "This deletes all lobbies, players and nations. Type 'yes' to continue: " ok
    [ "$ok" = "yes" ] || { echo "aborted"; exit 1; }
    just down
    docker volume rm ironlobby_pgdata

# install frontend dependencies if missing
_deps:
    @[ -d frontend/node_modules ] || (cd frontend && npm install)
