# Iron Lobby: frontend

Svelte 5 + Vite single-page app. No UI framework, no router library: three
routes and a hand-written dark "war room" theme (`src/app.css`).

```sh
npm install
npm run dev      # http://localhost:3001, proxies /api -> http://localhost:3000
npm run check    # svelte-check (types + a11y)
npm run build    # static files in dist/
```

Start the API first (`docker compose up` in the repo root). The dev server
proxies `/api` so the browser sees one origin: the auth cookie works without
CORS. Point it elsewhere with `IRONLOBBY_API=http://host:port npm run dev`.

## Routes

| Path         | What                                                        |
| ------------ | ----------------------------------------------------------- |
| `/`          | Host or join (paste invite link / ID), "Your lobbies" list  |
| `/new`       | Create a lobby: briefing, game mode, start time, nations    |
| `/lobby/:id` | Claim/switch/leave a nation; host edits slots, details, ID  |

## Deploying

`dist/` is static. Serve it so that `/api/*` reaches the Go server on the same
origin, and unknown paths fall back to `index.html` (client-side routing).
To call an API on another origin instead, build with `VITE_API_BASE=https://…/api`
(that needs CORS + `SameSite=None` cookies on the API side).

## Identity

The API identifies players by the `username` UUID cookie from `POST /api/user`
and has no "who am I" endpoint. The app stores the callsign and the IDs of
lobbies you host in `localStorage` (`ironlobby.v1`). Any action that needs a
player opens the callsign dialog first, or again if the server rejects the cookie.

## Layout

```
src/
  lib/        api client, identity, router, shared state, nation metadata, time utils
  components/ Modal, NationCard, NationPicker, Stepper, Tag, Toasts, EnlistModal, Icon
  routes/     Home, CreateLobby, Lobby, Missing
```
