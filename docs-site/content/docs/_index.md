---
title: Introduction
weight: 1
---

# Introduction

**mini-claude** is a terminal chat client for self-hosted LLMs, paired with a
small sync server. It is written in Go and built around three ideas:

1. **Privacy** — inference runs locally through [Ollama](https://ollama.com);
   nothing about your prompts leaves your machine unless you choose to sync.
2. **Local-first** — the CLI works fully offline. Your settings and conversation
   history are saved as JSON in your config directory.
3. **Optional sync** — create an account and the server backs up your config and
   history (PostgreSQL/JSONB), so you can restore them on any machine.

## The two halves

| Component | Role | Tech |
|---|---|---|
| **CLI (TUI)** | chat, settings, auth, local persistence | Go, Bubble Tea, Lip Gloss |
| **Server** | accounts, data sync, backups, metrics | Go, Gin, PostgreSQL, JWT |

Read on for [installation](/docs/installation/), the
[commands](/docs/usage/), and the [architecture](/docs/architecture/).
