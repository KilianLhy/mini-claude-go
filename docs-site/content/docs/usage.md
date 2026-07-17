---
title: Utilisation
weight: 3
---

# Utilisation

Lancer le client :

```bash
mini-claude
```

Tape un message et appuie sur `entrée` pour discuter. `ctrl+j` insère un saut de
ligne.

## Commandes

| Commande | Action |
|---|---|
| `/help` | affiche toutes les commandes |
| `/model` | choisir ou changer de modèle |
| `/settings` | changer le thème de couleurs |
| `/login` · `/register` | se connecter / créer un compte |
| `/logout` | se déconnecter |
| `/export` | pousser config + historique vers le serveur |
| `/import` | récupérer config + historique depuis le serveur |
| `/clear` | démarrer une nouvelle conversation |
| `/quit` | quitter |

## Configuration

Les réglages sont dans `~/.config/mini-claude/config.json`, et l'historique dans
`state.json`. Chaque valeur peut être surchargée par une variable
d'environnement :

| Variable | Défaut | Signification |
|---|---|---|
| `MINI_CLAUDE_URL` | `http://localhost:11434` | URL d'Ollama |
| `MINI_CLAUDE_MODEL` | `llama3.2:3b` | modèle par défaut |
| `MINI_CLAUDE_TEMPERATURE` | `0.7` | température d'échantillonnage |
| `MINI_CLAUDE_SYSTEM` | — | prompt système |
| `MINI_CLAUDE_THEME` | `claude` | thème de couleurs |
| `MINI_CLAUDE_SERVER` | URL hébergée | URL du serveur de synchro |

Priorité : **variable d'environnement > `config.json` > défaut intégré**.
