# Agents

## Agent Principal (`cmd/agent`)

**Rôle :** Tracking automatique de l'activité utilisateur

**Fonctionnalités :**
- Détecte la fenêtre active toutes les 2s
- Stocke les activités dans SQLite
- Détection d'inactivité (>60s)
- Serveur HTTP (port 8787) avec dashboard web et API REST

**Lancement :**
```bash
./trackmytime
```

## Agent Export (`cmd/export`)

**Rôle :** Export des données trackées

**Formats :** CSV, JSON

**Usage :**
```bash
./trackmytime-export                 # CSV today
./trackmytime-export -format json    # JSON today
./trackmytime-export -period week    # Semaine
./trackmytime-export -aggregated     # Données agrégées
```

---

**Base de données :** `~/.trackmytime/activities.db` (partagée entre les deux agents)
