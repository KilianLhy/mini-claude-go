---
title: API serveur
weight: 5
---

# API serveur

Le serveur de synchronisation est une API REST construite avec Gin. Les mots de
passe sont hashés avec bcrypt ; les routes `/me/*` exigent un en-tête
`Authorization: Bearer <jwt>`. La config et l'état sont stockés en JSONB
PostgreSQL.

## Endpoints

| Méthode & chemin | Auth | Description |
|---|---|---|
| `POST /auth/register` | — | créer un compte → JWT |
| `POST /auth/login` | — | se connecter → JWT |
| `GET /me/data` | JWT | lire la config + l'état courants |
| `PUT /me/data` | JWT | mettre à jour la config + l'état |
| `POST /me/export` | JWT | pousser les données et créer une sauvegarde |
| `POST /me/import` | JWT | récupérer les données courantes |
| `GET /me/backups` | JWT | lister les sauvegardes |
| `GET /me/backups/:id` | JWT | récupérer une sauvegarde |
| `GET /health` | — | vérification de santé |
| `GET /metrics` | — | métriques Prometheus |

## Modèle de données

```
users(id, email, password_hash, created_at)
user_data(user_id, config JSONB, state JSONB, updated_at)   -- courant, upsert
backups(id, user_id, config JSONB, state JSONB, created_at) -- historique
```

## Exemple

```bash
# inscription → token
curl -X POST https://hugostarte.alwaysdata.net/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"toi@example.com","password":"supersecret"}'
```
