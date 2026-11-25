# Spec: host player management (kick + move)

**Goal:** the host can remove any player from the lobby, and move any player to
another nation (the frontend does the move by drag and drop).

Both routes are host-only, go in the authorized `r.Group` in `routes.go`, and
start with `checkHost`, like the other host routes.

## 1. Kick a player

```
DELETE /api/lobby/{lobby_id}/player/{player_name}
```

- **Effect:** deletes that player's `player_lobby` row. The nation stays; the
  seat becomes free.
- **Body:** none.
- **Success:** `204 No Content` (or `200`; the frontend only checks for 2xx and
  then re-fetches the lobby).

| Status | When                                          | Body text                   |
| ------ | --------------------------------------------- | --------------------------- |
| 403    | requester isn't the host                      | `only the host can do that` |
| 404    | player not in this lobby / lobby doesn't exist | `player not found in lobby` |

**SQL** (`queries/lobby.sql`):

```sql
-- name: KickLobbyPlayer :one
DELETE FROM player_lobby
WHERE lobby_id = $1
  AND player_id = (SELECT id FROM player WHERE player_name = $2)
RETURNING *;
```

`pgx.ErrNoRows` → `ErrUserNotFound` → 404, same pattern as `DeleteLobbyPlayer`.

## 2. Move a player to another nation

```
PATCH /api/lobby/{lobby_id}/player/{player_name}
{ "country_tag": "ITA" }
```

- **Effect:** changes the player's nation and **keeps their note and
  `joined_at`**.
- **Slot limits:** same rules as joining. The target nation must be open in the
  lobby and have a free slot. Moving to the nation they already hold is a no-op
  success.
- **Success:** `200`.

| Status | When                          | Body text                                                       |
| ------ | ----------------------------- | --------------------------------------------------------------- |
| 400    | nation not open in lobby      | `requested nation is unavailable` (existing `ErrNationNotAvail`) |
| 400    | target nation is full         | `requested nation is full` (existing `ErrNationSlotsFull`)       |
| 403    | requester isn't the host      | `only the host can do that`                                     |
| 404    | player not in this lobby      | `player not found in lobby`                                     |

**SQL:**

```sql
-- name: GetPlayerIDByName :one
SELECT id FROM player WHERE player_name = $1;

-- name: MoveLobbyPlayer :one
UPDATE player_lobby
SET country_id = (SELECT id FROM countries WHERE country_tag = $3)
WHERE lobby_id = $1 AND player_id = $2
RETURNING *;
```

**Repo method** `MoveLobbyPlayer(ctx, lobbyID, playerName, tag)`, in one
transaction, mirroring `UpsertLobbyPlayer`:

1. `GetPlayerIDByName`: no rows → `ErrUserNotFound`.
2. `LockLobbyCountrySlot(lobbyID, tag)`: no rows → `ErrNationNotAvail`.
3. `CountCountryOccupants(lobbyID, tag, playerID)` ≥ max slots →
   `ErrNationSlotsFull`. This query already excludes the moved player, so a
   no-op move passes.
4. `MoveLobbyPlayer`: no rows → `ErrUserNotFound`, since the player exists but
   isn't in this lobby.
5. Commit.

## Things to watch

- **Use 403 with a message for "not host", not a bare 401.** The frontend
  treats a 401, or a 400 with an empty body, as "your session is gone" and opens
  the callsign dialog. `checkHost` currently answers non-hosts with a bare 401,
  which would trigger that by mistake. Fixing it inside `checkHost` also
  improves the existing host routes.
- **Player names in the URL:** the frontend will send
  `encodeURIComponent(name)`. Callsigns can contain spaces, `%` or `/`, so
  decode with `url.PathUnescape(chi.URLParam(r, "player_name"))`. chi keeps an
  encoded `%2F` inside a single segment.
- **Kick isn't a ban.** A kicked player can claim a slot again right away.
  Blocking that would need a ban table, which is out of scope here.
- The host kicking or moving **themselves** is fine and behaves like
  leave/switch.

## Suggested tests (`server_test.go`)

- Host kicks a player → 2xx, player gone from `GET /lobby/{id}`.
- Non-host tries to kick → 403, player still there.
- Kick a name that isn't in the lobby → 404.
- Move into a free slot → 200, nation changed, note unchanged.
- Move into a full nation → 400, nothing changed.
- Move to a nation that isn't open in the lobby → 400.

## Frontend (built once the routes exist)

- An **×** next to each player on the nation cards (host only), with a
  confirmation ("Remove Rundstedt from Germany?").
- **Drag and drop** of player rows between nation cards (host only). Full
  nations are dimmed while dragging and refuse the drop; a failed move shows the
  server's error and snaps back. On touch screens, a "Move to…" menu per player
  instead, since drag and drop is unreliable there.
