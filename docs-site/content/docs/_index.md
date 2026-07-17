---
title: Introduction
weight: 1
---

# Introduction

**mini-claude** est un client de chat en terminal pour LLM auto-hébergés,
associé à un petit serveur de synchronisation. Écrit en Go, il repose sur trois
idées :

1. **Confidentialité** — l'inférence tourne en local via
   [Ollama](https://ollama.com) ; rien de tes prompts ne quitte ta machine, sauf
   si tu choisis de synchroniser.
2. **Local d'abord** — la CLI fonctionne entièrement hors-ligne. Tes réglages et
   ton historique sont sauvegardés en JSON dans ton dossier de config.
3. **Synchro optionnelle** — crée un compte et le serveur sauvegarde ta config et
   ton historique (PostgreSQL/JSONB), pour les restaurer sur n'importe quelle machine.

## Les deux moitiés

| Composant | Rôle | Techno |
|---|---|---|
| **CLI (TUI)** | chat, réglages, auth, persistance locale | Go, Bubble Tea, Lip Gloss |
| **Serveur** | comptes, synchro, sauvegardes, métriques | Go, Gin, PostgreSQL, JWT |

Poursuis avec l'[installation](/docs/installation/), les
[commandes](/docs/usage/), et l'[architecture](/docs/architecture/).
