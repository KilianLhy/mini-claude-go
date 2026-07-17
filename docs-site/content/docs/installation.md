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

## Pre-built binaries

Download a binary for Windows, macOS or Linux from the
[GitHub releases](https://github.com/KilianLhy/mini-claude-go/releases).

## Prerequisite: Ollama

The CLI needs a running [Ollama](https://ollama.com) with a model pulled:

```bash
ollama pull llama3.2:3b
```

By default the CLI talks to Ollama on `http://localhost:11434`.

## From source

```bash
git clone https://gitlab.com/marseille-bb/mini-claude
cd mini-claude
go run ./cmd/tui
```
