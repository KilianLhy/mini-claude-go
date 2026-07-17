---
title: Déploiement
weight: 6
---

# Déploiement

Le serveur est un unique binaire Go qui parle à PostgreSQL. Il se configure
entièrement par variables d'environnement :

| Variable | Signification |
|---|---|
| `PORT` | port HTTP (fourni automatiquement par la plupart des hébergeurs) |
| `DATABASE_URL` | DSN PostgreSQL ; vide ⇒ stockage en mémoire (dev uniquement) |
| `JWT_SECRET` | secret pour signer les tokens d'authentification |

Le schéma est créé automatiquement au démarrage (migration idempotente).

## Cross-compilation pour Linux

```bash
make build-server-linux
# -> bin/mini-claude-server-linux-amd64  (binaire ELF statique)
```

## Hébergement (alwaysdata)

Le serveur de production tourne sur [alwaysdata](https://www.alwaysdata.com),
qui héberge à la fois le binaire Go et PostgreSQL sur une offre gratuite. En
résumé :

1. Créer une base PostgreSQL et noter sa chaîne de connexion.
2. Cross-compiler et uploader le binaire Linux en SSH.
3. Créer un site de type **User program** pointant sur le binaire, avec
   `DATABASE_URL` et `JWT_SECRET` définis en variables d'environnement.

L'instance en ligne répond sur `https://hugostarte.alwaysdata.net/health`.

## Local (Docker)

```bash
docker compose up -d --build
```

Démarre PostgreSQL, le serveur, Ollama et la stack de supervision en une seule
commande.
