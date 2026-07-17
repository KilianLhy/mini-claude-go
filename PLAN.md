# Plan d'action — mini-claude

Objectif : couvrir toute la grille de notation en gardant une archi **propre et
cohérente** avec ce qui existe déjà (CLI TUI + chat streaming vers Ollama) et avec la
stack annoncée dans le README (Go, Bubble Tea, `net/http`, Ollama, Docker, GitLab CI).

## Point de départ (déjà acquis)

- Écran d'accueil avec options — **0.5** ✅
- Écrans fonctionnalité principale (chat + sélecteur de modèle) — **1** ✅
- Interface agréable, colorée, dynamique — **1** ✅

**Total acquis : ~2.5 pts** sur ~15 pts techniques (hors soutenance 5 + démo 1).

---

## Principes d'architecture (à respecter partout)

1. **Offline-first.** La CLI marche à 100 % sans compte ni serveur : le chat local
   (Ollama) fonctionne toujours. Le serveur = couche de **sync optionnelle** (comptes +
   sauvegarde). Si le serveur est down, la CLI dégrade proprement, elle ne plante pas.
2. **Monorepo, un seul `go.mod`.** Deux binaires : `cmd/tui` (existant) et
   `cmd/server` (nouveau). Le module reste `github.com/KilianLhy/mini-claude-go`
   (pas de rename, `go install` continue de marcher).
3. **Un contrat partagé.** Les formats JSON échangés CLI↔serveur (config, state, auth)
   sont définis **une seule fois** dans un paquet `internal/shared` (ou `pkg/api`),
   importé des deux côtés. Impossible de diverger.
4. **Erreurs gérées partout.** Fichier absent / JSON corrompu / Ollama éteint /
   serveur injoignable → message clair, jamais de panic, on repart sur un défaut sain.

## Décision hébergement / distribution

Le repo reste **GitLab** (source de vérité, historique, branches, CI — cohérent avec le
README « GitLab CI/CD »). Le critère « release GitHub » est couvert par un **repo GitHub
de distribution uniquement** :

- GoReleaser tourne dans la **CI GitLab** sur les tags `v*`.
- Il publie la **release sur GitHub** (binaires) + met à jour les taps
  **Homebrew** et **Scoop** (hébergés sur GitHub, où c'est la convention).
- GitLab reste le dépôt de dev ; GitHub n'est qu'une vitrine de release. Une seule
  pipeline, une seule source de vérité.

## Décision produit

Ollama reste le **moteur d'inférence local** — on n'y touche pas. Notre serveur est un
**backend de comptes + sauvegarde** : il stocke, par utilisateur, la config de la CLI
(modèle préféré, thème, URL Ollama…), l'état (historique des conversations) et des
backups horodatés. Ça relie le chat existant à la partie serveur exigée sans hors-sujet.

Flux : `TUI ⇄ Ollama (local, inférence)` + `TUI ⇄ notre serveur (sync config/état)`.

---

## Phase 0 — Fondations (avant de coder les features)
**Effort : faible. Débloque le travail en parallèle.**

- Poser la structure monorepo : `cmd/tui`, `cmd/server`, `internal/shared`.
- Écrire le **contrat d'API** dans `internal/shared` : types `Config`, `State`,
  `RegisterRequest`, `LoginResponse`, etc. + la liste des routes (voir Phase 3).
- Se mettre d'accord à 3 sur ce contrat → Hugo (CLI) et Kilian (serveur) bossent en
  parallèle sans se bloquer.

## Phase 1 — Persistance JSON + écran paramètres (CLI seule)
**Points visés : +1.5** (config.json 0.5, state.json 0.5, écran paramètres 0.5)
**Effort : faible-moyen. Aucune dépendance serveur.**

- `internal/store/` : lecture/écriture de `~/.config/mini-claude/config.json` et
  `state.json` via `os.UserConfigDir()` (portable Win/Mac/Linux).
- `config.Load()` lit le JSON en plus des variables d'env (priorité :
  env > fichier > défauts).
- `state.json` : sérialiser l'historique de chat (`chat.History` existe déjà). Charger
  au démarrage, sauver à chaque échange / à la fermeture.
- Écran **Paramètres** (nouveau `mode` dans `ui.go`) : thème de couleurs (2-3 palettes),
  modèle par défaut, URL Ollama. Persisté dans config.json.
- Erreurs : fichier absent → créer avec défauts ; JSON corrompu → message + défauts.

## Phase 2 — Distribution (GoReleaser, indépendant du serveur)
**Points visés : +3** (cross-compile CLI Win/Linux/Mac 1, releases GitHub 1, package manager 1)
**Effort : faible. Fort ROI.**

- Créer le repo GitHub de distribution + les repos `homebrew-tap` et `scoop-bucket`.
- `.goreleaser.yaml` : builds `windows/linux/darwin` × `amd64/arm64`, release cible
  **GitHub**, sections `brews:` (Homebrew) et `scoops:` (Scoop).
- `.gitlab-ci.yml` : job `release` déclenché sur tag `v*` qui lance GoReleaser
  (token GitHub en variable CI protégée).
- Première release `v0.1.0` dès la Phase 1 mergée : l'artefact, c'est la CLI.

## Phase 3 — Serveur REST (Gin + PostgreSQL/JSONB)
**Points visés : +4** (auth 1, données 1, import/export 1, backups 1)
**Effort : élevé. Le cœur manquant.**

- `cmd/server/main.go` + `internal/api/` : **Gin** + **pgx** vers **PostgreSQL**,
  données en **JSONB**. Importe les DTO de `internal/shared`.
- Schéma (état courant séparé de l'historique) :
  - `users(id, email, password_hash, created_at)`
  - `user_data(user_id PK, config JSONB, state JSONB, updated_at)` — courant, upsert
  - `backups(id, user_id, config JSONB, state JSONB, created_at)` — historique horodaté
- Endpoints :
  - `POST /auth/register`, `POST /auth/login` → **JWT** (mot de passe hashé **bcrypt**).
  - `GET/PUT /me/data` — récupérer/mettre à jour config+state courants.
  - `POST /me/export` (push courant **et** crée un backup), `POST /me/import` (pull courant).
  - `GET /me/backups`, `GET /me/backups/:id` — lister/restaurer un backup.
- Middleware auth JWT sur `/me/*`. Erreurs : codes HTTP corrects + body `{ "error": ... }`.

## Phase 4 — Écran auth CLI + câblage sync (dépend de la Phase 3)
**Points visés : +1** (écran authentification 0.5, import/export vers serveur 0.5)
**Effort : moyen.**

- `internal/apiclient/` : appels vers notre serveur (register/login, get/put data,
  export/import, backups). JWT stocké dans `credentials.json` (perms 0600, non versionné).
- Écran **Authentification** (nouveau `mode`) : email + mot de passe, login/register.
- Depuis l'écran Paramètres : **Exporter vers le serveur**, **Importer depuis le
  serveur**, liste des backups. Tout reste optionnel (offline-first).

## Phase 5 — Docker, cross-compile serveur Linux, hébergement (dépend de la Phase 3)
**Points visés : +2** (cross-compile serveur Linux 1, hébergement gratuit 1)
**Effort : faible-moyen.**

- `docker-compose.yml` (dev) : Postgres + serveur. `Dockerfile` multi-stage pour le
  serveur (comble aussi le « Docker » annoncé dans le README).
- Cross-compile Linux : `GOOS=linux GOARCH=amd64 go build ./cmd/server` (ou via
  GoReleaser). Vérifier l'exécution sous Linux.
- Déployer sur **Render** (web service + Postgres gratuits) ou **alwaysdata**. DSN
  Postgres et secret JWT en variables d'env. URL publique = défaut (surchargeable) CLI.

## Phase 6 — Qualité, tests & livrables
**Sécurise démo (1) + soutenance (5) + robustesse.**

- **Tests** (`go test ./...` promis dans CONTRIBUTING, aujourd'hui vide) : round-trip
  JSON du `store`, précédence de config, handlers d'API (auth, export/import).
- Durcir la gestion d'erreurs (timeouts HTTP, serveur down, offline).
- `SELF-EVAL.md` : chaque critère de la grille + note attribuée + total (livrable obligatoire).
- README à jour (install Homebrew/Scoop, config serveur). Archive du code pour le dépôt scolaire.
- Prépa démo 10 min + questions techniques (ci-dessous).

---

## Récap points

| Phase | Contenu | Points |
|---|---|---|
| Acquis | accueil + fonctionnalité + interface | 2.5 |
| 0 | fondations (monorepo + contrat) | 0 (débloque) |
| 1 | persistance JSON + écran paramètres | +1.5 |
| 2 | GoReleaser : cross-compile CLI + releases GitHub + package manager | +3 |
| 3 | serveur Gin + Postgres (auth, data, import/export, backups) | +4 |
| 4 | écran auth CLI + câblage sync | +1 |
| 5 | Docker + cross-compile serveur Linux + hébergement | +2 |
| — | soutenance (5) + démo (1) | 6 |
| | **Total visé** | **~20** |

## Répartition suggérée (3 personnes)

- **Hugo** — CLI/TUI : Phase 1 (persistance + paramètres) puis Phase 4 (auth + sync).
  Continuité du code déjà écrit.
- **Kilian** — Serveur : Phase 3 (Gin + Postgres, endpoints, JWT) puis Phase 5
  (Docker, cross-compile Linux, Render).
- **Moustapha** — Infra/release : Phase 2 (GoReleaser, taps, CI GitLab) + tests +
  `SELF-EVAL.md` + prépa démo.

Prérequis commun : figer le **contrat d'API** (`internal/shared`) en Phase 0.

## Questions techniques probables en soutenance

- Bubble Tea / archi Elm (Model-Update-View), rôle des `tea.Msg` / `tea.Cmd`.
- Streaming SSE : pourquoi goroutine + channels + `select` sur `ctx.Done()` ?
- Pourquoi `Messages()` renvoie une copie du slice (bug corrigé `af4e76b`) ?
- Pourquoi JSONB plutôt que des colonnes classiques ? Index GIN ?
- Sécurité auth : hash bcrypt, JWT, expiration, stockage du token côté client.
- Cross-compilation : rôle de `GOOS`/`GOARCH`, ce que GoReleaser fait en plus.
- Gestion d'erreurs : Ollama éteint ? serveur down ? JSON corrompu ? (offline-first)
- Pourquoi monorepo + paquet de contrat partagé plutôt que deux repos ?
