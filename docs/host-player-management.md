# Request: host player management (kick + move)

The host needs to be able to remove any player from the lobby, and move any
player to another nation (the frontend will do the move by drag and drop).

## 1. Kick a player

```
DELETE /api/lobby/{lobby_id}/player/{player_name}
```

- Host only.
- Removes that player's seat. The nation stays open and the seat becomes free.
- A kicked player can claim a slot again right away; no ban needed.
- The host kicking themselves is fine and works like leaving.
- No request body. On success, any 2xx; I re-fetch the lobby afterwards.

| Status | When                                            | Body              |
| ------ | ----------------------------------------------- | ----------------- |
| 403    | requester isn't the host                        | a short message   |
| 404    | player isn't in this lobby, or no such lobby    | a short message   |

## 2. Move a player to another nation

```
PATCH /api/lobby/{lobby_id}/player/{player_name}
{ "country_tag": "ITA" }
```

- Host only.
- Changes the player's nation and keeps their note and join time.
- Same slot rules as a player claiming a nation: the target nation must be
  open in the lobby and have a free slot.
- Moving a player to the nation they already hold should succeed and change
  nothing.
- On success, any 2xx.

| Status | When                          | Body            |
| ------ | ----------------------------- | --------------- |
| 400    | nation isn't open in lobby    | a short message |
| 400    | target nation is full         | a short message |
| 403    | requester isn't the host      | a short message |
| 404    | player isn't in this lobby    | a short message |

## Applies to both

- Error bodies are shown to the user as-is, so keep them short and readable.
- Please don't answer "not host" with a bare 401. The frontend reads a 401 (or
  a 400 with an empty body) as "your session is gone" and opens the callsign
  prompt.
- `{player_name}` arrives URL-encoded (`encodeURIComponent`). Callsigns can
  contain spaces, `%` or `/`.

## What I'll build once the routes exist

- An × next to each player on the nation cards (host only), with a
  confirmation.
- Drag and drop of players between nation cards (host only). Full nations
  refuse the drop; a failed move shows your error message and snaps back. On
  touch screens, a "Move to…" menu per player instead.
