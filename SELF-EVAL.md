# Auto-évaluation — mini-claude

Évaluation de notre travail au regard de la grille de notation du projet.

## CLI en mode TUI

| Critère | Max | Note | Justification |
|---|---|---|---|
| Écran d'accueil avec les options | 0.5 | 0.5 | Écran de bienvenue (logo ASCII, liste des commandes, infos modèle/serveur) — `internal/ui/ui.go` (`welcomeView`). |
| Écran d'authentification | 0.5 | 0.5 | Écran `/login` / `/register` (email + mot de passe masqué, bascule login/register) — `internal/ui/auth.go`. |
| Écran de paramètres (thème) | 0.5 | 0.5 | Écran `/settings` avec 3 palettes et aperçu live — `internal/ui/theme.go`. |
| Écrans de la fonctionnalité principale | 1 | 1 | Chat en streaming avec un LLM local + sélecteur de modèle — `internal/ui/ui.go`, `internal/client`. |
| Stockage des paramètres en JSON (`~/.config`) | 0.5 | 0.5 | `~/.config/mini-claude/config.json`, écriture atomique — `internal/store`. |
| Stockage de l'état en JSON (`~/.config`) | 0.5 | 0.5 | `~/.config/mini-claude/state.json` (historique) — `internal/store`. |
| Import/export config + état vers le serveur | 0.5 | 0.5 | Commandes `/export` et `/import` — `internal/apiclient`. |
| Interface agréable, colorée, dynamique | 1 | 1 | Lip Gloss, thèmes, spinner, streaming token par token, liens OSC 8. |

**Sous-total CLI : 5 / 5**

## Serveur d'API REST

| Critère | Max | Note | Justification |
|---|---|---|---|
| Endpoints d'authentification | 1 | 1 | `POST /auth/register`, `POST /auth/login` → JWT ; mots de passe hashés bcrypt — `internal/api`. |
| Endpoints données (get/update) | 1 | 1 | `GET` / `PUT /me/data`, protégés par middleware JWT. |
| Endpoints import/export | 1 | 1 | `POST /me/export`, `POST /me/import`. |
| Stockage des backups par utilisateur | 1 | 1 | Table `backups` (PostgreSQL, JSONB), `GET /me/backups`, `GET /me/backups/:id`. |

**Sous-total serveur : 4 / 4**

## Distribution & hébergement

| Critère | Max | Note | Justification |
|---|---|---|---|
| Cross-compiler le serveur pour Linux | 1 | 1 | `make build-server-linux` (binaire ELF Linux amd64 statique). |
| Héberger le serveur gratuitement | 1 | 1 | Déployé sur alwaysdata : `https://hugostarte.alwaysdata.net` (Postgres managé). |
| Cross-compiler la CLI (Win/Linux/macOS) | 1 | 1 | GoReleaser — 6 cibles (windows/linux/darwin × amd64/arm64). |
| Publier dans les releases GitHub | 1 | 1 | Releases sur `github.com/KilianLhy/mini-claude-go`. |
| Publier sur un gestionnaire de paquets | 1 | 1 | Homebrew : `brew tap KilianLhy/tap && brew install mini-claude`. |

**Sous-total distribution : 5 / 5**

## Soutenance

| Critère | Max | Note | Justification |
|---|---|---|---|
| Répondre aux questions techniques | 5 | — | À évaluer en soutenance. Choix techniques documentés et maîtrisés (voir README + commentaires du code). |
| Qualité de la démonstration | 1 | — | À évaluer en soutenance. Scénario de démo préparé. |

## Contraintes respectées

- **Gestion des erreurs** : fichiers absents/corrompus (repli sur défauts), erreurs HTTP (timeouts, serveur injoignable, codes 4xx/5xx avec corps JSON), la CLI reste utilisable hors-ligne.
- **Travail en binôme/trinôme** : Hugo Stawiarski, Kilian Lahaye, Moustapha Sow.
- **Pas de plagiat** ; usage de l'IA assumé, choix techniques justifiables.

## Bonus (hors barème)

- **Observabilité** : endpoint `/metrics` (client Prometheus), scraping Prometheus, et **dashboard Grafana provisionné automatiquement** (métriques applicatives : inscriptions, connexions, exports/imports, latences par route ; + métriques conteneurs via cAdvisor/node-exporter).
- **Tests** : suites pour le store, l'API et le client HTTP (`go test ./...`).
- **CI/CD** : pipeline GitLab (lint/build/test + release GoReleaser sur tag).

## Total

- **Points techniques (hors soutenance/démo) : 14 / 14**
- **Total projeté (avec soutenance réussie) : 20 / 20**
