# Nakama Profile Module

A Go runtime module for [Nakama](https://heroiclabs.com/docs/nakama/) that lets players save a profile. It runs on Nakama 3.41.0 and Postgres 16 with Docker Compose.

## Getting started

You need Docker with Docker Compose. To build the module and start the server, run:

```bash
docker compose up --build
```

| Port | Service |
|---|---|
| 7349 | gRPC API |
| 7350 | HTTP API |
| 7351 | Nakama Console |

## RPCs

| RPC | Description |
|---|---|
| `healthcheck` | Returns `{"success": true}` |
| `save_profile` | Validates and saves the authenticated user's profile |

### `save_profile`

Example payload:

```json
{"display_name": "PlayerOne", "avatar": "knight_01", "bio": "Hello there!"}
```

Validation rules:

- `display_name` must be 3–32 printable characters after surrounding whitespace is trimmed. It can't contain `mod`, `admin`, or `system`, in any letter case.
- `bio` can be up to 190 characters.
- `avatar` isn't validated yet.

Profiles are stored in the `profiles` collection under the key `profile`. Anyone can read them, but clients can't write to them directly.
