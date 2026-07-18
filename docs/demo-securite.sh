#!/usr/bin/env bash
#
# Démo sécurité mini-claude — à lancer pendant la soutenance.
#
# Prérequis : dans un AUTRE terminal, lance le serveur en local :
#     go run ./cmd/server
# (sans base de données → store mémoire, secret de dev. Écoute sur :8080)
#
# Puis exécute ce script :  bash docs/demo-securite.sh
# ou lance les blocs un par un pour commenter à l'oral.

BASE="http://localhost:8080"

line() { echo; echo "======================================================"; echo "$1"; echo "======================================================"; }

# ---------------------------------------------------------------------------
line "0. Happy path — le service fonctionne normalement"
# On crée un compte et on récupère un token JWT (réutilisé plus bas).
TOKEN=$(curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"supersecret"}' \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Token obtenu : ${TOKEN:0:24}..."
echo "GET /me/data avec le token :"
curl -s -o /dev/null -w "  → %{http_code} (attendu 200)\n" \
  "$BASE/me/data" -H "Authorization: Bearer $TOKEN"

# ---------------------------------------------------------------------------
line "1. Rate limiting — anti brute-force"
# 12 logins ratés d'affilée : 401 puis bascule en 429.
for i in $(seq 1 12); do
  curl -s -o /dev/null -w "  tentative $i → %{http_code}\n" \
    -X POST "$BASE/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"x@y.com","password":"whatever12"}'
done
echo "Attendu : 401 puis 429 à partir de la ~6e."

# ---------------------------------------------------------------------------
line "2. Limite de taille de requête — anti-DoS"
# Corps de 3 Mo (> 2 Mo) => refusé.
head -c 3000000 /dev/zero | tr '\0' 'a' > /tmp/big.txt
curl -s -o /dev/null -w "  gros corps (3 Mo) → %{http_code} (attendu 400)\n" \
  -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  --data-binary @/tmp/big.txt
rm -f /tmp/big.txt

# ---------------------------------------------------------------------------
line "3. En-têtes de sécurité"
curl -is "$BASE/health" | grep -iE 'X-Content-Type|X-Frame|Referrer'
echo "Attendu : nosniff, DENY, no-referrer."

# ---------------------------------------------------------------------------
line "4. Authentification obligatoire sur /me/*"
echo "Sans token :"
curl -s -o /dev/null -w "  → %{http_code} (attendu 401)\n" "$BASE/me/data"
echo "Avec un faux token :"
curl -s -o /dev/null -w "  → %{http_code} (attendu 401)\n" \
  "$BASE/me/data" -H "Authorization: Bearer pas-un-vrai-jwt"

# ---------------------------------------------------------------------------
line "5. Énumération de comptes bloquée (même 401)"
echo "Email inconnu :"
curl -s -w "  → %{http_code} : " -o /dev/null \
  -X POST "$BASE/auth/login" -H "Content-Type: application/json" \
  -d '{"email":"inconnu@x.com","password":"supersecret"}'; echo
echo "Bon email, mauvais mot de passe :"
curl -s -w "  → %{http_code} : " -o /dev/null \
  -X POST "$BASE/auth/login" -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"mauvais"}'; echo
echo "Attendu : 401 identique dans les deux cas (on ne révèle pas quels emails existent)."

# ---------------------------------------------------------------------------
line "6. Injection SQL via le contenu d'un message"
# On stocke un message contenant une injection classique, puis on relit.
curl -s -o /dev/null -w "  export payload malveillant → %{http_code} (attendu 201)\n" \
  -X POST "$BASE/me/export" -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"config":{},"state":{"messages":[{"role":"user","content":"'"'"'; DROP TABLE users;--"}]}}'
echo "Relecture (import) — la donnée revient intacte, la table users existe toujours :"
curl -s "$BASE/me/import" -X POST -H "Authorization: Bearer $TOKEN" | head -c 300; echo
echo "Requêtes paramétrées => l'injection est traitée comme du texte."

# ---------------------------------------------------------------------------
line "7. Isolation entre comptes (autorisation)"
# User A pousse un secret, user B ne doit PAS le voir.
TA=$(curl -s -X POST "$BASE/auth/register" -H "Content-Type: application/json" \
  -d '{"email":"userA@x.com","password":"supersecret"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
curl -s -o /dev/null -X POST "$BASE/me/export" -H "Authorization: Bearer $TA" \
  -H "Content-Type: application/json" \
  -d '{"config":{"model":"SECRET-DE-A"},"state":{"messages":[]}}'
TB=$(curl -s -X POST "$BASE/auth/register" -H "Content-Type: application/json" \
  -d '{"email":"userB@x.com","password":"supersecret"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "User B lit ses données (ne doit PAS contenir SECRET-DE-A) :"
curl -s "$BASE/me/import" -X POST -H "Authorization: Bearer $TB" | head -c 300; echo
echo "Chaque requête filtre sur le userID du JWT, jamais sur un paramètre client."

# ---------------------------------------------------------------------------
line "8. Secret JWT robuste imposé en production"
echo "Ces deux commandes sont à lancer à la main (elles bloquent le démarrage) :"
echo
echo "  # Prod (DATABASE_URL) SANS secret => refuse de démarrer :"
echo '  DATABASE_URL="postgres://fake" go run ./cmd/server'
echo "  → 'JWT_SECRET must be set to at least 32 characters in production'"
echo
echo "  # Avec un secret >= 32 caractères => passe la vérif :"
echo '  DATABASE_URL="postgres://fake" JWT_SECRET="0123456789012345678901234567890123" go run ./cmd/server'

# ---------------------------------------------------------------------------
line "9. Tests automatiques (sans serveur)"
echo "  go test ./internal/api -v -run 'RateLimit|Login|Register|Auth'"
echo
echo "Fin de la démo."
