# Déploiement du serveur sur alwaysdata

Le serveur `mini-claude` est un binaire Go unique qui écoute en HTTP et se
connecte à PostgreSQL. alwaysdata (offre gratuite) héberge **les deux** — le
binaire et la base — dans un seul compte.

Résumé : on **cross-compile** le serveur en binaire Linux, on l'**uploade**, on
crée une **base PostgreSQL**, on configure un **site** qui lance le binaire avec
les bonnes variables d'environnement.

## 1. Créer le compte

Créer un compte gratuit sur <https://www.alwaysdata.com>. Noter le **nom de
compte** (`ACCOUNT`) : il apparaît dans les URLs (`ACCOUNT.alwaysdata.net`) et
les chemins SSH (`/home/ACCOUNT/`).

## 2. Créer la base PostgreSQL

Dans le panneau : **Databases → PostgreSQL → Add a database**.

- Nom de la base : `miniclaude`
- Créer un utilisateur + mot de passe (les noter)
- Hôte : en général `postgresql-ACCOUNT.alwaysdata.net`, port `5432`

Construire la chaîne de connexion (`DATABASE_URL`) :

```
postgres://USER:PASSWORD@postgresql-ACCOUNT.alwaysdata.net:5432/miniclaude?sslmode=disable
```

> Le serveur crée les tables tout seul au démarrage (migration idempotente),
> rien à faire à la main côté schéma.

## 3. Cross-compiler le serveur

Sur ton Mac, depuis la racine du repo :

```bash
make build-server-linux
```

Ça produit `bin/mini-claude-server-linux-amd64` : un binaire **Linux x86-64
statique** (aucune dépendance à installer sur le serveur).

## 4. Uploader le binaire

Via SFTP/SSH (identifiants SSH dans le panneau alwaysdata, **Remote access →
SSH**) :

```bash
ssh ACCOUNT@ssh-ACCOUNT.alwaysdata.net "mkdir -p ~/mini-claude"
scp bin/mini-claude-server-linux-amd64 ACCOUNT@ssh-ACCOUNT.alwaysdata.net:~/mini-claude/server
ssh ACCOUNT@ssh-ACCOUNT.alwaysdata.net "chmod +x ~/mini-claude/server"
```

## 5. Créer le site (programme)

Dans le panneau : **Sites → Add a site → Type = "User program" (programme
utilisateur)**.

- **Command** : `/home/ACCOUNT/mini-claude/server`
- **Adresse/port** : alwaysdata attribue un port à ton programme. Le serveur
  écoute sur `:$PORT` (variable d'environnement). Renseigne le port du site et
  mets la variable `PORT` sur la même valeur (étape 6).
- Domaine : associer `ACCOUNT.alwaysdata.net` (ou un sous-domaine).

## 6. Variables d'environnement

Toujours dans la config du site (ou **Environment**), définir :

| Variable | Valeur |
|---|---|
| `DATABASE_URL` | la chaîne de l'étape 2 |
| `JWT_SECRET` | un secret long et aléatoire — génère-le avec `openssl rand -hex 32` |
| `PORT` | le port attribué au site |

> `JWT_SECRET` doit rester secret et stable (s'il change, tous les tokens
> existants deviennent invalides). Ne jamais le committer.

## 7. Démarrer et vérifier

Redémarrer le site depuis le panneau, puis :

```bash
curl https://ACCOUNT.alwaysdata.net/health      # -> {"status":"ok"}
```

Tester le parcours complet :

```bash
curl -X POST https://ACCOUNT.alwaysdata.net/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"supersecret"}'
```

Doit renvoyer un `token`.

## 8. Pointer la CLI vers le serveur en ligne

Deux options :

- **Ponctuel** : lancer la CLI avec `MINI_CLAUDE_SERVER=https://ACCOUNT.alwaysdata.net go run ./cmd/tui`
- **Par défaut** (recommandé pour la démo) : mettre l'URL comme valeur par
  défaut dans `internal/config/config.go` (`DefaultServerURL`), puis refaire une
  release de la CLI. Ainsi le correcteur n'a rien à configurer.

## Dépannage

- **502 / site ne répond pas** : le programme n'écoute pas sur le bon port —
  vérifier que `PORT` (env) == port du site, et que le binaire est exécutable.
- **`ping postgres` / erreur de connexion** : vérifier `DATABASE_URL`
  (utilisateur, mot de passe, hôte). Essayer `sslmode=require` si `disable`
  échoue.
- **Logs** : consultables dans le panneau alwaysdata (**Logs**) ou via SSH.
