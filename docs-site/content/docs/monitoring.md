---
title: Supervision
weight: 7
---

# Supervision

Le serveur expose des métriques applicatives sur `/metrics` grâce au client Go
de Prometheus. La stack Docker locale embarque une supervision complète :

- **Prometheus** scrape le serveur, plus node-exporter (machine) et cAdvisor
  (conteneurs).
- **Grafana** provisionne automatiquement une source de données et un tableau de
  bord **mini-claude**.

## Métriques exposées

| Métrique | Signification |
|---|---|
| `miniclaude_http_requests_total` | requêtes par méthode, route, statut |
| `miniclaude_http_request_duration_seconds` | histogramme des latences |
| `miniclaude_auth_events_total` | inscriptions / connexions (succès & échec) |
| `miniclaude_sync_events_total` | exports / imports |

## Accès

Avec la stack lancée (`docker compose up -d`) :

- Grafana → <http://localhost:3000> (admin / admin) → tableau de bord
  **mini-claude — application & infra**
- Prometheus → <http://localhost:9090> (`/targets` pour vérifier l'état du scrape)

Les panneaux applicatifs se remplissent dès qu'il y a du trafic sur le serveur ;
les panneaux CPU/mémoire des conteneurs se remplissent tout seuls via cAdvisor.
