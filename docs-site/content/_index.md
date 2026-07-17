---
title: mini-claude
type: docs
---

# mini-claude

Un client de chat rapide et privé pour LLM auto-hébergés, en Go — accompagné
d'un serveur de synchronisation pour que tes conversations et réglages te
suivent d'une machine à l'autre.

Discute avec un modèle de langage local depuis ton terminal. L'inférence reste
sur ta machine (via [Ollama](https://ollama.com)) ; un compte optionnel te
permet de sauvegarder et restaurer ta config et ton historique depuis ton
propre serveur.

## Points forts

- **Interface terminal** (Bubble Tea) — réponses en streaming, chat multi-tours, thèmes.
- **Local d'abord** — réglages et historique stockés en JSON dans ton dossier de config.
- **Compte & synchro** — connecte-toi, puis pousse/récupère tes données vers un serveur auto-hébergé.
- **Serveur REST** — Gin + PostgreSQL (JSONB), authentification JWT, sauvegardes par utilisateur.
- **Supervision** — métriques Prometheus + un tableau de bord Grafana provisionné.

[**Commencer →**](/docs/installation/)

---

Réalisé par Hugo Stawiarski · Kilian Lahaye · Moustapha Sow — licence MIT.
