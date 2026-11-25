# Request: delete lobby

The host needs to be able to delete their lobby. The button is already built
(Edit lobby → Delete lobby → confirm).

## What I need

```
DELETE /api/lobby/{lobby_id}
```

- Host only.
- Deletes the lobby for everyone: its nations and every player's seat go with
  it. Afterwards `GET /api/lobby/{lobby_id}` returns 404.
- No request body.
- On success, any 2xx. I don't read the response body.

## Errors the frontend handles

| Status | When                     | Body                                     |
| ------ | ------------------------ | ---------------------------------------- |
| 403    | requester isn't the host | a short message, shown to the user as-is |
| 404    | lobby doesn't exist      | a short message, shown to the user as-is |

Please don't answer "not host" with a bare 401. The frontend reads a 401 (or a
400 with an empty body) as "your session is gone" and opens the callsign
prompt.
