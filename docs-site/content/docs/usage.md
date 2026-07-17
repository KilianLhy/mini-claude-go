---
title: Usage
weight: 3
---

# Usage

Launch the client:

```bash
mini-claude
```

Type a message and press `enter` to chat. Use `ctrl+j` to insert a newline.

## Commands

| Command | Action |
|---|---|
| `/help` | show all commands |
| `/model` | pick or switch the model |
| `/settings` | change the color theme |
| `/login` · `/register` | sign in / create an account |
| `/logout` | sign out |
| `/export` | push config + history to the server |
| `/import` | pull config + history from the server |
| `/clear` | start a fresh conversation |
| `/quit` | exit |

## Configuration

Settings live in `~/.config/mini-claude/config.json`, and history in
`state.json`. Any value can be overridden by an environment variable:

| Variable | Default | Meaning |
|---|---|---|
| `MINI_CLAUDE_URL` | `http://localhost:11434` | Ollama base URL |
| `MINI_CLAUDE_MODEL` | `llama3.2:3b` | default model |
| `MINI_CLAUDE_TEMPERATURE` | `0.7` | sampling temperature |
| `MINI_CLAUDE_SYSTEM` | — | system prompt |
| `MINI_CLAUDE_THEME` | `claude` | color theme |
| `MINI_CLAUDE_SERVER` | hosted URL | sync server URL |

Precedence: **environment variable > `config.json` > built-in default**.
