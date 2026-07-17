---
title: Server API
weight: 5
---

# Server API

The sync server is a REST API built with Gin. Passwords are hashed with bcrypt;
`/me/*` routes require an `Authorization: Bearer <jwt>` header. Config and state
are stored as PostgreSQL JSONB.

## Endpoints

| Method & path | Auth | Description |
|---|---|---|
| `POST /auth/register` | — | create an account → JWT |
| `POST /auth/login` | — | sign in → JWT |
| `GET /me/data` | JWT | read current config + state |
| `PUT /me/data` | JWT | update current config + state |
| `POST /me/export` | JWT | push data and create a backup |
| `POST /me/import` | JWT | pull current data |
| `GET /me/backups` | JWT | list backups |
| `GET /me/backups/:id` | JWT | fetch one backup |
| `GET /health` | — | health check |
| `GET /metrics` | — | Prometheus metrics |

## Data model

```
users(id, email, password_hash, created_at)
user_data(user_id, config JSONB, state JSONB, updated_at)   -- current, upsert
backups(id, user_id, config JSONB, state JSONB, created_at) -- history
```

## Example

```bash
# register → token
curl -X POST https://hugostarte.alwaysdata.net/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"supersecret"}'
```
