# Iron Lobby

A small application for Heart of Iron 4 to help people organize multiplayer lobbies!

## Motivation

Hoi4 multiplayer games tend to last a long time (up to 6 hours) and are most the fun
with lots of players! This poses a challenge when organizing matches. Lobbies
have to be planned in advance, with player having to negotiate various parameter
such as game rules and which nations are being played.

IronLobby aims to simplify this process by providing an easy to use online
interface for hosting lobbies, assigning nations, distributing the invite code
and more!

## Design Goals

- The App has two modes: hosting a game or joining a game
- When hosting, the user gets host privileges, such as defining game rules,
assigning nations and distributing the invite ID for when the game is supposed
to commence
- When joining a lobby, users can reserve/queue/or otherwise signal their
preference for a certain nation

## Implementation

## API routes

```

UNAUTHORIZED ROUTES:

GET /{lobby_id} -> ruft Lobby auf

POST /user -> register new user
json "username":"{user_input}"

AUTHORIZED ROUTES:

POST /lobby -> neue lobby erstellen
body:
type InsertLobbyParams struct {
        LobbyName  string    `json:"lobby_name"`
        StartsAt   time.Time `json:"starts_at"`
        Gamemode   Gamemode  `json:"gamemode"`
}

Host actions:

PUT /{lobby_id} -> lobby bearbeiten
body:
type InsertLobbyParams struct {
        LobbyName  string    `json:"lobby_name"`
        StartsAt   time.Time `json:"starts_at"`
        Gamemode   Gamemode  `json:"gamemode"`
}

DELETE /{lobby_id} -> lobby löschen

Player actions:

POST /{lobby_id} -> lobby joinen
body:
type AssignPlayerToLobbyParams struct {
        CountryTag string    `json:"country_tag"`
}

PUT /edit/{lobby_id} -> nation wechseln/lobby leaven?
```
### Running the App

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
