---
title: Installation
weight: 2
---

# Installation

## Homebrew (macOS / Linux)

```bash
brew tap KilianLhy/tap
brew install mini-claude
mini-claude
```

## Binaires pré-compilés

Télécharge un binaire pour Windows, macOS ou Linux depuis les
[releases GitHub](https://github.com/KilianLhy/mini-claude-go/releases).

## Prérequis : Ollama

La CLI a besoin d'un [Ollama](https://ollama.com) qui tourne, avec un modèle
téléchargé :

```bash
ollama pull llama3.2:3b
```

Par défaut, la CLI parle à Ollama sur `http://localhost:11434`.

## Depuis les sources

```bash
git clone https://github.com/KilianLhy/mini-claude-go
cd mini-claude-go
go run ./cmd/tui
```
