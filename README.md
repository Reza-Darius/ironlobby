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

### How to run

requires docker

```
git clone https://github.com/Reza-Darius/ironlobby
docker compose up
```

iron Lobby is configured via an `.env` file with the following values:

```
PORT=3000
DEBUG_CORS=true

POSTGRES_USER=ironlobby-user
POSTGRES_PASSWORD=oberkommando
POSTGRES_DB=ironlobby-db
```

### API routes

look up `routes.go` to see the API routes
