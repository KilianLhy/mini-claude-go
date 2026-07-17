---
title: Architecture
weight: 4
---

# Architecture

La CLI fonctionne entièrement hors-ligne (le chat passe par Ollama en local). Le
serveur est une couche de synchronisation optionnelle : connecte-toi pour
sauvegarder et restaurer tes données.

```
 Client TUI (Go)                     Serveur de synchro (Go)      Stockage
┌────────────────┐   HTTP/JSON     ┌──────────────────┐        ┌────────────┐
│  Bubble Tea    │ ───────────────▶│  API REST Gin    │───────▶│ PostgreSQL │
│  chat, auth,   │   (auth JWT)    │  auth, données,  │  JSONB │ users/data │
│  réglages      │◀─────────────── │  export, backups │        │ backups    │
└───────┬────────┘                 └────────┬─────────┘        └────────────┘
        │ compatible OpenAI                 │ /metrics
        ▼ HTTP (local)                       ▼
   ┌──────────┐                      ┌──────────────┐
   │  Ollama  │                      │ Prometheus + │
   │  (local) │                      │  Grafana     │
   └──────────┘                      └──────────────┘
```

C'est **la CLI** qui parle à Ollama, directement — le serveur ne fait jamais
d'inférence. Ça garde les prompts en local (confidentialité), permet le mode
hors-ligne, et laisse le serveur assez léger pour être hébergé gratuitement.

## Organisation du code

```
cmd/tui/          point d'entrée de la CLI
cmd/server/       point d'entrée du serveur
internal/ui/      modèle Bubble Tea, vues, thèmes, écran d'auth
internal/client/  client de streaming Ollama
internal/chat/    historique de conversation
internal/config/  chargement de la configuration
internal/store/   persistance JSON locale (config, état, identifiants)
internal/apiclient/ client HTTP vers le serveur de synchro
internal/api/     serveur Gin : handlers, auth, stockage, métriques
internal/shared/  le contrat JSON partagé par la CLI et le serveur
```

## Choix de conception

- **Une seule source de vérité pour le format d'échange** : `internal/shared`
  définit les types JSON importés à la fois par la CLI et le serveur — ils ne
  peuvent donc jamais diverger.
- **Stockage interchangeable** : le serveur parle à une interface `Store`,
  implémentée par un backend PostgreSQL (production) et un backend en mémoire
  (tests / zéro configuration).
- **Local d'abord** : un fichier local absent ou corrompu retombe sur les
  valeurs par défaut ; la CLI ne plante jamais si le serveur ou Ollama est
  injoignable.
